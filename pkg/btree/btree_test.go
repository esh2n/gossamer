// Copyright 2023 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package btree

import (
	"testing"

	"github.com/ChainSafe/gossamer/pkg/scale"

	"github.com/stretchr/testify/require"
)

type dummy struct {
	Field1 uint32
	Field2 [32]byte
}

func TestBTree_Codec(t *testing.T) {
	comparator := func(a, b interface{}) bool {
		v1 := a.(dummy)
		v2 := b.(dummy)
		return v1.Field1 < v2.Field1
	}

	// Create a Tree with 3 dummy items
	tree := NewTree[dummy](comparator)
	tree.Set(dummy{Field1: 1})
	tree.Set(dummy{Field1: 2})
	tree.Set(dummy{Field1: 3})
	encoded, err := scale.Marshal(tree)
	require.NoError(t, err)

	//let mut btree = Map::<u32, Hash>::new();
	//btree.insert(1, Hash::zero());
	//btree.insert(2, Hash::zero());
	//btree.insert(3, Hash::zero());
	//let encoded = btree.encode();
	//println!("encoded: {:?}", encoded);
	expectedEncoded := []byte{12,
		1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	require.Equal(t, expectedEncoded, encoded)

	expected := NewTree[dummy](comparator)
	err = scale.Unmarshal(expectedEncoded, &expected)
	require.NoError(t, err)

	// Check that the expected Tree has the same items as the original
	require.Equal(t, tree.Len(), expected.Len())
	require.Equal(t, tree.ItemType, expected.ItemType)
	require.Equal(t, tree.Min(), expected.Min())
	require.Equal(t, tree.Max(), expected.Max())
	require.Equal(t, tree.Get(dummy{Field1: 1}), expected.Get(dummy{Field1: 1}))
	require.Equal(t, tree.Get(dummy{Field1: 2}), expected.Get(dummy{Field1: 2}))
	require.Equal(t, tree.Get(dummy{Field1: 3}), expected.Get(dummy{Field1: 3}))
}

func TestBTreeMap_Codec(t *testing.T) {
	btreeMap := NewMap[uint32, dummy](32)
	btreeMap.Set(uint32(1), dummy{Field1: 1})
	btreeMap.Set(uint32(2), dummy{Field1: 2})
	btreeMap.Set(uint32(3), dummy{Field1: 3})
	encoded, err := scale.Marshal(btreeMap)
	require.NoError(t, err)

	//let mut btree = Map::<u32, (u32, Hash)>::new();
	//btree.insert(1, (1, Hash::zero()));
	//btree.insert(2, (2, Hash::zero()));
	//btree.insert(3, (3, Hash::zero()));
	//let encoded = btree.encode();
	//println!("encoded: {:?}", encoded);
	expectedEncoded := []byte{12, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	require.Equal(t, expectedEncoded, encoded)
	expected := NewMap[uint32, dummy](32)
	err = scale.Unmarshal(expectedEncoded, &expected)
	require.NoError(t, err)
	require.Equal(t, btreeMap.Len(), expected.Len())
}