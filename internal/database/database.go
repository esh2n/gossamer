package database

import (
	"io"
)

type Reader interface {
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
}

type Writer interface {
	Put(key, value []byte) error
	Del(key []byte) error
	Flush() error
}

// Iterator iterates over key/value pairs in ascending key order.
// Must be released after use.
type Iterator interface {
	Valid() bool
	Next() bool
	Key() []byte
	Value() []byte
	Release()
}

type PrefixIterator interface {
}

// Batch is a write-only operation.
type Batch interface {
	Writer

	ValueSize() int
	Reset()
}

// Database wraps all database operations. All methods are safe for concurrent use.
type Database interface {
	Reader
	Writer
	io.Closer

	Path() string
	NewBatch() Batch
	NewIterator() Iterator
	NewPrefixIterator(prefix []byte) Iterator
}

type Table interface {
	Reader
	Writer
	io.Closer
	Path() string
	NewBatch() Batch
	NewIterator() Iterator
}