// Copyright 2023 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package btree

import (
	"fmt"
	"io"
	"reflect"

	"github.com/ChainSafe/gossamer/pkg/scale"

	"golang.org/x/exp/constraints"

	"github.com/tidwall/btree"
)

type Codec interface {
	MarshalSCALE() ([]byte, error)
	UnmarshalSCALE(reader io.Reader) error
}

// Tree is a wrapper around tidwall/btree.BTree that also stores the comparator function and the type of the items
// stored in the BTree. This is needed during decoding because the Tree item is a generic type, and we need to know it
// at the time of decoding.
type Tree struct {
	*btree.BTree
	Comparator func(a, b interface{}) bool
	ItemType   reflect.Type
}

// MarshalSCALE encodes the Tree using SCALE.
func (bt Tree) MarshalSCALE() ([]byte, error) {
	encodedLen, err := scale.Marshal(uint(bt.Len()))
	if err != nil {
		return nil, fmt.Errorf("failed to encode BTree length: %w", err)
	}

	var encodedItems []byte
	bt.Ascend(nil, func(item interface{}) bool {
		var encodedItem []byte
		encodedItem, err = scale.Marshal(item)
		if err != nil {
			return false
		}

		encodedItems = append(encodedItems, encodedItem...)
		return true
	})

	return append(encodedLen, encodedItems...), err
}

// UnmarshalSCALE decodes the Tree using SCALE.
func (bt Tree) UnmarshalSCALE(reader io.Reader) error {
	if bt.Comparator == nil {
		return fmt.Errorf("comparator not found")
	}

	sliceType := reflect.SliceOf(bt.ItemType)
	slicePtr := reflect.New(sliceType)
	encodedItems, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read BTree items: %w", err)
	}
	err = scale.Unmarshal(encodedItems, slicePtr.Interface())
	if err != nil {
		return fmt.Errorf("decode BTree items: %w", err)
	}

	for i := 0; i < slicePtr.Elem().Len(); i++ {
		item := slicePtr.Elem().Index(i).Interface()
		bt.Set(item)
	}
	return nil
}

// Copy returns a copy of the Tree.
func (bt Tree) Copy() *Tree {
	return &Tree{
		BTree:      bt.BTree.Copy(),
		Comparator: bt.Comparator,
		ItemType:   bt.ItemType,
	}
}

// NewTree creates a new Tree with the given comparator function.
func NewTree[T any](comparator func(a, b any) bool) Tree {
	elementType := reflect.TypeOf((*T)(nil)).Elem()
	return Tree{
		BTree:      btree.New(comparator),
		Comparator: comparator,
		ItemType:   elementType,
	}
}

var _ Codec = (*Tree)(nil)

// Map is a wrapper around tidwall/btree.Map
type Map[K constraints.Ordered, V any] struct {
	*btree.Map[K, V]
	Degree int
}

type mapItem[K constraints.Ordered, V any] struct {
	Key   K
	Value V
}

// MarshalSCALE encodes the Map using SCALE.
func (btm Map[K, V]) MarshalSCALE() ([]byte, error) {
	encodedLen, err := scale.Marshal(uint(btm.Len()))
	if err != nil {
		return nil, fmt.Errorf("failed to encode Map length: %w", err)
	}

	var (
		pivot        K
		encodedItems []byte
	)
	btm.Ascend(pivot, func(key K, value V) bool {
		var (
			encodedKey   []byte
			encodedValue []byte
		)
		encodedKey, err = scale.Marshal(key)
		if err != nil {
			return false
		}

		encodedValue, err = scale.Marshal(value)
		if err != nil {
			return false
		}

		encodedItems = append(encodedItems, encodedKey...)
		encodedItems = append(encodedItems, encodedValue...)
		return true
	})

	return append(encodedLen, encodedItems...), err
}

// UnmarshalSCALE decodes the Map using SCALE.
func (btm Map[K, V]) UnmarshalSCALE(reader io.Reader) error {
	if btm.Degree == 0 {
		return fmt.Errorf("nothing to decode into")
	}

	if btm.Map == nil {
		btm.Map = btree.NewMap[K, V](btm.Degree)
	}

	sliceType := reflect.SliceOf(reflect.TypeOf((*mapItem[K, V])(nil)).Elem())
	slicePtr := reflect.New(sliceType)
	encodedItems, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read Map items: %w", err)
	}
	err = scale.Unmarshal(encodedItems, slicePtr.Interface())
	if err != nil {
		return fmt.Errorf("decode Map items: %w", err)
	}

	for i := 0; i < slicePtr.Elem().Len(); i++ {
		item := slicePtr.Elem().Index(i).Interface().(mapItem[K, V])
		btm.Map.Set(item.Key, item.Value)
	}
	return nil
}

// Copy returns a copy of the Map.
func (btm Map[K, V]) Copy() Map[K, V] {
	return Map[K, V]{
		Map: btm.Map.Copy(),
	}
}

// NewMap creates a new Map with the given degree.
func NewMap[K constraints.Ordered, V any](degree int) Map[K, V] {
	return Map[K, V]{
		Map:    btree.NewMap[K, V](degree),
		Degree: degree,
	}
}

var _ Codec = (*Map[int, string])(nil)