// Copyright 2024 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package api

import (
	"slices"

	"github.com/ChainSafe/gossamer/internal/primitives/core/hash"
	"github.com/ChainSafe/gossamer/internal/primitives/database"
	"github.com/ChainSafe/gossamer/internal/primitives/runtime"
	"github.com/ChainSafe/gossamer/pkg/scale"
	"github.com/tidwall/btree"
)

type leafSetItem[H comparable, N runtime.Number] struct {
	hash   H
	number N
}

// ImportOutcome privately contains the inserted and removed leaves after an import action.
type ImportOutcome[H comparable, N runtime.Number] struct {
	inserted leafSetItem[H, N]
	removed  *H
}

// RemoveOutcome privates contains the inserted and removed leaves after a remove action.
type RemoveOutcome[H comparable, N runtime.Number] struct {
	inserted *H
	removed  leafSetItem[H, N]
}

// FinalizationOutcome contains the removed leaves after a finalization action.
type FinalizationOutcome[H comparable, N runtime.Number] struct {
	removed btree.Map[N, []H]
}

// Leaves returns the leaves that were removed after a finalization action.
func (fo FinalizationOutcome[H, N]) Leaves() []H {
	leaves := make([]H, 0)
	for _, hashes := range fo.removed.Values() {
		leaves = append(leaves, hashes...)
	}
	return leaves
}

// list of leaf hashes ordered by number (descending).
// stored in memory for fast access.
// this allows very fast checking and modification of active leaves.
type LeafSet[H comparable, N runtime.Number] struct {
	storage btree.Map[N, []H]
}

// NewLeafSet is constructor for a new, blank `LeafSet`.
func NewLeafSet[H comparable, N runtime.Number]() LeafSet[H, N] {
	return LeafSet[H, N]{
		storage: *btree.NewMap[N, []H](0),
	}
}

// NewLeafSetFromDB will read the leaf list from the DB, using given prefix for keys.
func NewLeafSetFromDB[H comparable, N runtime.Number](
	db database.Database[hash.H256], column uint32, prefix []byte,
) (LeafSet[H, N], error) {
	storage := btree.NewMap[N, []H](0)

	leaves := db.Get(database.ColumnID(column), prefix)
	if leaves != nil {
		type numberHashes struct {
			Number N
			Hashes []H
		}
		var vals []numberHashes
		err := scale.Unmarshal(leaves, &vals)
		if err != nil {
			return LeafSet[H, N]{}, err
		}
		for _, nh := range vals {
			storage.Set(nh.Number, nh.Hashes)
		}
	}
	return LeafSet[H, N]{
		storage: *storage,
	}, nil
}

// Import will update the leaf list on import.
func (ls *LeafSet[H, N]) Import(hash H, number N, parentHash H) ImportOutcome[H, N] {
	var removed *H
	if number != 0 {
		parentNumber := number - 1
		if ls.removeLeaf(parentNumber, parentHash) {
			removed = &parentHash
		}
	}

	ls.insertLeaf(number, hash)
	return ImportOutcome[H, N]{
		inserted: leafSetItem[H, N]{hash, number},
		removed:  removed,
	}
}

// Remove will update the leaf list on removal.
//
// Note that the leaves set structure doesn't have the information to decide if the
// leaf we're removing is the last children of the parent. Follows that this method requires
// the caller to check this condition and optionally pass the `parentHash` if `hash` is
// its last child.
//
// Returns `nil` if no modifications are applied.
func (ls *LeafSet[H, N]) Remove(hash H, number N, parentHash *H) *RemoveOutcome[H, N] {
	if !ls.removeLeaf(number, hash) {
		return nil
	}

	var inserted *H
	if parentHash != nil {
		if number != 0 {
			parentNumber := number - 1
			ls.insertLeaf(parentNumber, *parentHash)
			inserted = parentHash
		} else {
			return nil
		}
	}

	return &RemoveOutcome[H, N]{
		inserted: inserted,
		removed:  leafSetItem[H, N]{hash, number},
	}
}

// FinalizeHeight will note a block height finalized, displacing all leaves with number less than the finalized
// block's.
//
// Although it would be more technically correct to also prune out leaves at the
// same number as the finalized block, but with different hashes, the current behavior
// is simpler and our assumptions about how finalization works means that those leaves
// will be pruned soon afterwards anyway.
func (ls *LeafSet[H, N]) FinalizeHeight(number N) FinalizationOutcome[H, N] {
	if number == 0 {
		removed := btree.NewMap[N, []H](0)
		return FinalizationOutcome[H, N]{removed: *removed}
	}
	boundary := number - 1
	belowBoundary := btree.NewMap[N, []H](0)
	ls.storage.Ascend(boundary, func(key N, value []H) bool {
		belowBoundary.Set(key, value)
		ls.storage.Delete(key)
		return false
	})
	return FinalizationOutcome[H, N]{removed: *belowBoundary}
}

// DisplacedByFinalHeight is the same as `FinalizeHeight()`, but it only simulates the operation.
//
// This means that no changes are done.
//
// Returns the leaves that would be displaced by finalizing the given block.
func (ls *LeafSet[H, N]) DisplacedByFinalHeight(number N) FinalizationOutcome[H, N] {
	if number == 0 {
		removed := btree.NewMap[N, []H](0)
		return FinalizationOutcome[H, N]{removed: *removed}
	}
	boundary := number - 1
	belowBoundary := btree.NewMap[N, []H](0)
	ls.storage.Ascend(boundary, func(key N, value []H) bool {
		belowBoundary.Set(key, value)
		return false
	})
	return FinalizationOutcome[H, N]{removed: *belowBoundary}
}

// Undo all pending operations.
//
// This returns an `Undo` struct, where any
// outcomes objects that have returned by previous method calls
// should be passed to via the appropriate methods. Otherwise,
// the on-disk state may get out of sync with in-memory state.
func (ls *LeafSet[H, N]) Undo() Undo[H, N] {
	return Undo[H, N]{ls}
}

// Revert to the given block height by dropping all leaves in the leaf set
// with a block number higher than the target.
func (ls *LeafSet[H, N]) Revert(bestHash H, bestNumber N) {
	items := make([]leafSetItem[H, N], 0)
	ls.storage.Reverse(func(number N, hashes []H) bool {
		for _, h := range hashes {
			items = append(items, leafSetItem[H, N]{h, number})
		}
		return true
	})

	for _, hn := range items {
		if hn.number > bestNumber {
			if !ls.removeLeaf(hn.number, hn.hash) {
				panic("item comes from an iterator over storage; qed")
			}
		}
	}

	var leavesContainBest bool
	hashes, ok := ls.storage.Get(bestNumber)
	if ok {
		leavesContainBest = slices.Contains(hashes, bestHash)
	}

	// we need to make sure that the best block exists in the leaf set as
	// this is an invariant of regular block import.
	if !leavesContainBest {
		ls.insertLeaf(bestNumber, bestHash)
	}
}

// Hashes returns a slice of all hashes in the leaf set
// ordered by their block number descending.
func (ls *LeafSet[H, N]) Hashes() []H {
	collected := make([]H, 0)
	ls.storage.Reverse(func(number N, hashes []H) bool {
		collected = append(collected, hashes...)
		return true
	})
	return collected
}

// Count is the number of known leaves.
func (ls *LeafSet[H, N]) Count() uint {
	var sum uint
	for _, level := range ls.storage.Values() {
		sum += uint(len(level))
	}
	return sum
}

// PrepareTransaction will write the leaf list to the database transaction.
func (ls *LeafSet[H, N]) PrepareTransaction(tx *database.Transaction[hash.H256], column uint32, prefix []byte) {
	type numberHashes struct {
		Number N
		Hashes []H
	}
	leaves := make([]numberHashes, 0)
	ls.storage.Reverse(func(number N, hashes []H) bool {
		leaves = append(leaves, numberHashes{number, hashes})
		return true
	})
	tx.Set(database.ColumnID(column), prefix, scale.MustMarshal(leaves))
}

// Contains will check if given block is a leaf.
func (ls *LeafSet[H, N]) Contains(number N, hash H) bool {
	hashes, ok := ls.storage.Get(number)
	if ok {
		return slices.Contains(hashes, hash)
	}
	return false
}

func (ls *LeafSet[H, N]) insertLeaf(number N, hash H) {
	hashes, ok := ls.storage.Get(number)
	if !ok {
		ls.storage.Set(number, []H{hash})
	} else {
		hashes = append(hashes, hash)
		ls.storage.Set(number, hashes)
	}
}

// Returns true if a leaf was found, false otherwise.
func (ls *LeafSet[H, N]) removeLeaf(number N, hash H) bool {
	var empty bool
	var removed bool
	leaves, ok := ls.storage.Get(number)
	if ok {
		var found bool
		retained := make([]H, 0)
		for _, h := range leaves {
			if h == hash {
				found = true
			} else {
				retained = append(retained, h)
			}
		}
		ls.storage.Set(number, retained)

		if len(retained) == 0 {
			empty = true
		}

		removed = found
	}

	if removed && empty {
		ls.storage.Delete(number)
	}

	return removed
}

// HighestLeaf returns the highest leaf and all hashes associated to it.
func (ls *LeafSet[H, N]) HighestLeaf() *struct {
	Number N
	Hashes []H
} {
	number, hashes, ok := ls.storage.Max()
	if !ok {
		return nil
	}
	return &struct {
		Number N
		Hashes []H
	}{
		Number: number,
		Hashes: hashes,
	}
}

// Undo is a helper for undoing operations.
type Undo[H comparable, N runtime.Number] struct {
	inner *LeafSet[H, N]
}

// UndoImport will undo an imported block by providing the import operation outcome.
// No additional operations should be performed between import and undo.
func (u Undo[H, N]) UndoImport(outcome ImportOutcome[H, N]) {
	if outcome.removed != nil {
		removedNumber := outcome.inserted.number - 1
		u.inner.insertLeaf(removedNumber, *outcome.removed)
	}
	u.inner.removeLeaf(outcome.inserted.number, outcome.inserted.hash)
}

// UndoRemove will undo a removed block by providing the displaced leaf.
// No additional operations should be performed between remove and undo.
func (u Undo[H, N]) UndoRemove(outcome RemoveOutcome[H, N]) {
	if outcome.inserted != nil {
		insertedNumber := outcome.removed.number - 1
		u.inner.removeLeaf(insertedNumber, *outcome.inserted)
	}
	u.inner.insertLeaf(outcome.removed.number, outcome.removed.hash)
}

// UndoFinalization will undo a finalization operation by providing the displaced leaves.
// No additional operations should be performed between finalization and undo.
func (u Undo[H, N]) UndoFinalization(outcome FinalizationOutcome[H, N]) {
	outcome.removed.Reverse(func(number N, hashes []H) bool {
		u.inner.storage.Set(number, hashes)
		return true
	})
}
