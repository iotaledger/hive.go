//go:build rocksdb
// +build rocksdb

package test

var (
	dbImplementations = []string{"badger", "bolt", "mapDB", "pebble", "rocksdb"}
)
