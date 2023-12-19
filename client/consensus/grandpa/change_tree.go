// Copyright 2023 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package grandpa

import (
	"errors"
	"fmt"

	"github.com/ChainSafe/gossamer/pkg/scale"
	"golang.org/x/exp/constraints"
)

/*
	The grandpa ChangeTree is a structure built to track pending changes across forks for the Grandpa Protocol.
	This structure is intended to represent an acyclic directed graph where the children are
    placed in descending order and number, you can ensure node ancestry using the `isDescendantOfFunc`.
*/

var (
	errDuplicateHashes     = errors.New("duplicated hashes")
	errUnfinalisedAncestor = errors.New("finalised descendent of Tree node without finalising its " +
		"ancestor(s) first")
	errRevert = errors.New("tried to import or finalise node that is an ancestor of " +
		"a previously finalised node")
)

// ChangeTree keeps track of the changes per fork allowing
// n forks in the same structure. This structure is intended
// to represent an acyclic directed graph where the hashNumber children are
// placed by descendency order and number, you can ensure an
// node ancestry using the `isDescendantOfFunc`
type ChangeTree[H comparable, N constraints.Unsigned, ID AuthorityID] struct {
	TreeRoots           []*PendingChangeNode[H, N, ID]
	BestFinalizedNumber *N
}

// NewChangeTree create an empty ChangeTree
func NewChangeTree[H comparable, N constraints.Unsigned, ID AuthorityID]() ChangeTree[H, N, ID] {
	return ChangeTree[H, N, ID]{}
}

// PendingChangeNode Represents a node in the ChangeTree
type PendingChangeNode[H comparable, N constraints.Unsigned, ID AuthorityID] struct {
	Change   *PendingChange[H, N, ID]
	Children []*PendingChangeNode[H, N, ID]
}

// Roots returns the roots of each fork in the ChangeTree
// This is the equivalent of the slice in the outermost layer of the roots
func (ct *ChangeTree[H, N, ID]) Roots() []*PendingChangeNode[H, N, ID] { //skipcq: RVV-B0011
	return ct.TreeRoots
}

// Import a new node into the roots.
//
// The given function `is_descendent_of` should return `true` if the second
// hash (target) is a descendent of the first hash (base).
//
// This method assumes that children in the same branch are imported in order.
//
// Returns `true` if the imported node is a root.
// WARNING: some users of this method (i.e. consensus epoch changes roots) currently silently
// rely on a **post-order DFS** traversal. If we are using instead a top-down traversal method
// then the `is_descendent_of` closure, when used after a warp-sync, may end up querying the
// backend for a block (the one corresponding to the root) that is not present and thus will
// return a wrong result.
func (ct *ChangeTree[H, N, ID]) Import(hash H,
	number N,
	change PendingChange[H, N, ID],
	isDescendentOf IsDescendentOf[H]) (bool, error) {
	for _, root := range ct.TreeRoots {
		imported, err := root.importNode(hash, number, change, isDescendentOf)
		if err != nil {
			return false, err
		}

		if imported {
			logger.Debugf("changes on header %s (%d) imported successfully",
				hash, number)
			return false, nil
		}
	}

	pendingChangeNode := &PendingChangeNode[H, N, ID]{
		Change: &change,
	}

	ct.TreeRoots = append(ct.TreeRoots, pendingChangeNode)
	return true, nil
}

// PendingChanges does a preorder traversal of the ChangeTree to get all pending changes
func (ct *ChangeTree[H, N, ID]) PendingChanges() []PendingChange[H, N, ID] {
	if len(ct.TreeRoots) == 0 {
		return nil
	}

	changes := make([]PendingChange[H, N, ID], 0, len(ct.TreeRoots))

	for i := 0; i < len(ct.TreeRoots); i++ {
		getPreOrder(&changes, ct.TreeRoots[i])
	}

	return changes
}

// getPreOrderChangeNodes does a preorder traversal of the ChangeTree to get all pending changes
func (ct *ChangeTree[H, N, ID]) getPreOrderChangeNodes() []*PendingChangeNode[H, N, ID] {
	if len(ct.TreeRoots) == 0 {
		return nil
	}

	changes := &[]*PendingChangeNode[H, N, ID]{}

	for i := 0; i < len(ct.TreeRoots); i++ {
		getPreOrderChangeNodes(changes, ct.TreeRoots[i])
	}

	return *changes
}

// FinalizesAnyWithDescendentIf Checks if any node in the tree is finalized by either finalising the
// node itself or a node's descendent that's not in the tree, guaranteeing
// that the node being finalized isn't a descendent of (or equal to) any of
// the node's children. Returns *true if the node being finalized is
// a root, *false if the node being finalized is not a root, and
// nil if no node in the tree is finalized. The given `Predicate` is
// checked on the prospective finalized root and must pass for finalisation
// to occur. The given function `is_descendent_of` should return `true` if
// the second hash (target) is a descendent of the first hash (base). func(T) bool
func (ct *ChangeTree[H, N, ID]) FinalizesAnyWithDescendentIf(
	hash *H,
	number N,
	isDescendentOf IsDescendentOf[H],
	predicate func(*PendingChange[H, N, ID]) bool) (*bool, error) {
	if ct.BestFinalizedNumber != nil {
		if number <= *ct.BestFinalizedNumber {
			return nil, errRevert
		}
	}

	roots := ct.Roots()

	nodes := ct.getPreOrderChangeNodes()

	// check if the given hash is equal or a descendent of any node in the
	// tree, if we find a valid node that passes the Predicate then we must
	// ensure that we're not finalising past any of its child nodes.
	for i := 0; i < len(nodes); i++ {
		root := nodes[i]
		isDesc, err := isDescendentOf(root.Change.CanonHash, *hash)
		if err != nil {
			return nil, err
		}

		if predicate(root.Change) && (root.Change.CanonHash == *hash || isDesc) {
			children := root.Children
			for _, child := range children {
				isChildDescOf, err := isDescendentOf(child.Change.CanonHash, *hash)
				if err != nil {
					return nil, err
				}

				if child.Change.CanonHeight <= number && (child.Change.CanonHash == *hash || isChildDescOf) {
					return nil, errUnfinalisedAncestor
				}
			}

			isEqual := false
			for _, val := range roots {
				if val.Change.CanonHash == root.Change.CanonHash {
					isEqual = true
					break
				}
			}
			return &isEqual, nil
		}
	}

	return nil, nil
}

// FinalizationResult Result of finalising a node (that could be a part of the roots or not).
type FinalizationResult scale.VaryingDataType

// Set will set a VaryingDataTypeValue using the underlying VaryingDataType
func (fr *FinalizationResult) Set(val scale.VaryingDataTypeValue) (err error) {
	vdt := scale.VaryingDataType(*fr)
	err = vdt.Set(val)
	if err != nil {
		return
	}
	*fr = FinalizationResult(vdt)
	return
}

// Value will return value from underying VaryingDataType
func (fr *FinalizationResult) Value() (val scale.VaryingDataTypeValue, err error) {
	vdt := scale.VaryingDataType(*fr)
	return vdt.Value()
}

func newFinalizationResult[H comparable, N constraints.Unsigned, ID AuthorityID]() FinalizationResult {
	vdt, err := scale.NewVaryingDataType(changed[H, N, ID]{}, unchanged{})
	if err != nil {
		panic(err)
	}
	return FinalizationResult(vdt)
}

type changed[H comparable, N constraints.Unsigned, ID AuthorityID] struct {
	value *PendingChange[H, N, ID]
}

func (changed[H, N, ID]) Index() uint {
	return 0
}

type unchanged struct{}

func (unchanged) Index() uint {
	return 1
}

// FinalizeWithDescendentIf Finalize a root in the roots by either finalising the node itself or a
// node's descendent that's not in the roots, guaranteeing that the node
// being finalized isn't a descendent of (or equal to) any of the root's
// children. The given `Predicate` is checked on the prospective finalized
// root and must pass for finalisation to occur. The given function
// `is_descendent_of` should return `true` if the second hash (target) is a
// descendent of the first hash (base).
func (ct *ChangeTree[H, N, ID]) FinalizeWithDescendentIf(
	hash *H,
	number N,
	isDescendentOf IsDescendentOf[H],
	predicate func(*PendingChange[H, N, ID]) bool) (result FinalizationResult, err error) {
	if ct.BestFinalizedNumber != nil {
		if number <= *ct.BestFinalizedNumber {
			return result, errRevert
		}
	}

	roots := ct.Roots()

	// check if the given hash is equal or a descendent of any root, if we
	// find a valid root that passes the Predicate then we must ensure that
	// we're not finalising past any children node.
	var position *N
	for i, root := range roots {
		isDesc, err := isDescendentOf(root.Change.CanonHash, *hash)
		if err != nil {
			return result, err
		}

		if predicate(root.Change) && (root.Change.CanonHash == *hash || isDesc) {
			for _, child := range root.Children {
				isDesc, err := isDescendentOf(child.Change.CanonHash, *hash)
				if err != nil {
					return result, err
				}
				if child.Change.CanonHeight <= number && (child.Change.CanonHash == *hash || isDesc) {
					return result, errUnfinalisedAncestor
				}
			}
			uintI := N(i)
			position = &uintI
			break
		}
	}

	var nodeData *PendingChange[H, N, ID]
	if position != nil {
		node := ct.swapRemove(ct.Roots(), *position)
		ct.TreeRoots = node.Children
		ct.BestFinalizedNumber = &node.Change.CanonHeight
		nodeData = node.Change
	}

	// Retain only roots that are descendents of the finalized block (this
	// happens if the node has been properly finalized) or that are
	// ancestors (or equal) to the finalized block (in this case the node
	// wasn't finalized earlier presumably because the Predicate didn't
	// pass).
	didChange := false
	roots = ct.Roots()

	ct.TreeRoots = []*PendingChangeNode[H, N, ID]{}
	for _, root := range roots {
		retain := false
		if root.Change.CanonHeight > number {
			isDescA, err := isDescendentOf(*hash, root.Change.CanonHash)
			if err != nil {
				return result, err
			}

			if isDescA {
				retain = true
			}
		} else if root.Change.CanonHeight == number && root.Change.CanonHash == *hash {
			retain = true
		} else {
			isDescB, err := isDescendentOf(root.Change.CanonHash, *hash)
			if err != nil {
				return result, err
			}

			if isDescB {
				retain = true
			}
		}
		if retain {
			ct.TreeRoots = append(ct.TreeRoots, root)
		} else {
			didChange = true
		}

		ct.BestFinalizedNumber = &number
	}

	result = newFinalizationResult[H, N, ID]()

	if nodeData != nil {
		err = result.Set(changed[H, N, ID]{
			value: nodeData,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	} else {
		if didChange {
			err = result.Set(changed[H, N, ID]{})
			if err != nil {
				return result, err
			}
			return result, nil
		} else {
			err = result.Set(unchanged{})
			if err != nil {
				return result, err
			}
			return result, nil
		}
	}
}

func (pcn *PendingChangeNode[H, N, ID]) importNode(hash H,
	number N,
	change PendingChange[H, N, ID],
	isDescendentOf IsDescendentOf[H]) (bool, error) {
	announcingHash := pcn.Change.CanonHash
	if hash == announcingHash {
		return false, fmt.Errorf("%w: %v", errDuplicateHashes, hash)
	}

	isDescendant, err := isDescendentOf(announcingHash, hash)
	if err != nil {
		return false, fmt.Errorf("cannot check ancestry: %w", err)
	}

	if !isDescendant {
		return false, nil
	}

	if number <= pcn.Change.CanonHeight {
		return false, nil
	}

	for _, childrenNodes := range pcn.Children {
		imported, err := childrenNodes.importNode(hash, number, change, isDescendentOf)
		if err != nil {
			return false, err
		}

		if imported {
			return true, nil
		}
	}
	childrenNode := &PendingChangeNode[H, N, ID]{
		Change: &change,
	}
	pcn.Children = append(pcn.Children, childrenNode)
	return true, nil
}

func getPreOrder[H comparable, N constraints.Unsigned, ID AuthorityID](
	changes *[]PendingChange[H, N, ID],
	changeNode *PendingChangeNode[H, N, ID]) {
	if changeNode == nil {
		return
	}

	if changes != nil {
		tempChanges := *changes
		tempChanges = append(tempChanges, *changeNode.Change)
		*changes = tempChanges
	} else {
		change := []PendingChange[H, N, ID]{*changeNode.Change}
		changes = &change
	}

	for i := 0; i < len(changeNode.Children); i++ {
		getPreOrder(changes, changeNode.Children[i])
	}
}

func getPreOrderChangeNodes[H comparable, N constraints.Unsigned, ID AuthorityID](
	changes *[]*PendingChangeNode[H, N, ID],
	changeNode *PendingChangeNode[H, N, ID]) {
	if changeNode == nil {
		return
	}

	if changes != nil {
		tempChanges := *changes
		tempChanges = append(tempChanges, changeNode)
		*changes = tempChanges
	} else {
		change := []*PendingChangeNode[H, N, ID]{changeNode}
		changes = &change
	}

	for i := 0; i < len(changeNode.Children); i++ {
		getPreOrderChangeNodes(changes, changeNode.Children[i])
	}
}

// Removes an element from the vector and returns it.
//
// The removed element is replaced by the last element of the vector.
//
// This does not preserve ordering, but is *O*(1).
//
// Panics if `index` is out of bounds.
func (ct *ChangeTree[H, N, ID]) swapRemove(roots []*PendingChangeNode[H, N, ID], index N) PendingChangeNode[H, N, ID] {
	if index >= N(len(roots)) {
		panic("swap_remove index out of bounds")
	}

	val := PendingChangeNode[H, N, ID]{}
	if roots[index] != nil {
		val = *roots[index]
	} else {
		panic("nil pending hashNumber node")
	}

	lastElem := roots[len(roots)-1]

	newRoots := roots[:len(roots)-1]
	// This should be the case where last elem was removed
	if index == N(len(newRoots)) {
		ct.TreeRoots = newRoots
		return val
	}
	newRoots[index] = lastElem
	ct.TreeRoots = newRoots
	return val
}

// Remove from the tree some nodes (and their subtrees) using a `filter` predicate.
//
// The `filter` is called over tree nodes and returns a filter action:
// - `Remove` if the node and its subtree should be removed;
// - `KeepNode` if we should maintain the node and keep processing the tree.
// - `KeepTree` if we should maintain the node and its entire subtree.
//
// An iterator over all the pruned nodes is returned.
func (_ *ChangeTree[H, N, ID]) drainFilter() { //nolint //skipcq: SCC-U1000 //skipcq: RVV-B0013
	// TODO implement
}