package boltdb

import (
	"errors"
	"go.etcd.io/bbolt"
	"math"

	"github.com/iotaledger/hive.go/database"
	"github.com/iotaledger/hive.go/objectstorage"
	"github.com/iotaledger/hive.go/objectstorage/boltdb"
)

const (
	BoltDBSubBatchSize = 10000
)

type BoltDB struct {
	bolt objectstorage.Storage
}

func NewDBWithPrefix(prefix []byte, db *bbolt.DB) *BoltDB {
	return &BoltDB{
		bolt: boltdb.New(db).WithRealm(prefix),
	}
}

// Read
func (db *BoltDB) Contains(key database.Key) (bool, error) {
	return db.bolt.Has(key)
}

func (db *BoltDB) Get(key database.Key) (database.Entry, error) {

	var entry database.Entry
	value, err := db.bolt.Get(key)
	if err != nil {
		if errors.Is(err, objectstorage.ErrKeyNotFound) {
			return entry, database.ErrKeyNotFound
		}
		return entry, err
	}
	entry.Key = key
	entry.Value = value
	return entry, nil
}

// Write
func (db *BoltDB) Set(entry database.Entry) error {
	return db.bolt.Set(entry.Key, entry.Value)
}

func (db *BoltDB) Delete(key database.Key) error {
	err := db.bolt.Delete(key)
	if err != nil {
		if errors.Is(err, objectstorage.ErrKeyNotFound) {
			return database.ErrKeyNotFound
		}
		return err
	}
	return nil
}

func (db *BoltDB) DeletePrefix(keyPrefix database.KeyPrefix) error {

	batch := db.bolt.Batched()
	err := db.ForEachPrefixKeyOnly(keyPrefix, func(key database.Key) bool {
		batch.Delete(key)
		return false
	})
	if err != nil {
		return err
	}
	return batch.Commit()
}

// Iteration
func (db *BoltDB) ForEach(consume func(entry database.Entry) (stop bool)) error {

	return db.bolt.Iterate([][]byte{}, true, func(key []byte, value []byte) bool {
		// Invert return value due to the difference in interfaces
		return !consume(database.Entry{
			Key:   key,
			Value: value,
		})
	})
}

func (db *BoltDB) ForEachPrefix(keyPrefix database.KeyPrefix, consume func(entry database.Entry) (stop bool)) error {

	return db.bolt.Iterate([][]byte{keyPrefix}, true, func(key []byte, value []byte) bool {
		// Invert return value due to the difference in interfaces
		return !consume(database.Entry{
			Key:   key,
			Value: value,
		})
	})

}

func (db *BoltDB) ForEachPrefixKeyOnly(keyPrefix database.KeyPrefix, consume func(entry database.Key) (stop bool)) error {
	return db.bolt.IterateKeys([][]byte{keyPrefix}, func(key []byte) bool {
		// Invert return value due to the difference in interfaces
		return !consume(key)
	})
}

func (db *BoltDB) StreamForEach(consume func(entry database.Entry) error) error {
	return db.ForEach(func(entry database.Entry) (stop bool) {
		return consume(entry) != nil
	})
}

func (db *BoltDB) StreamForEachKeyOnly(consume func(key database.Key) error) error {
	return db.bolt.IterateKeys([][]byte{}, func(key []byte) bool {
		// Invert return value due to the difference in interfaces
		return consume(key) == nil
	})
}

func (db *BoltDB) StreamForEachPrefix(keyPrefix database.KeyPrefix, consume func(entry database.Entry) error) error {
	return db.ForEachPrefix(keyPrefix, func(entry database.Entry) (stop bool) {
		return consume(entry) != nil
	})
}

func (db *BoltDB) StreamForEachPrefixKeyOnly(keyPrefix database.KeyPrefix, consume func(entry database.Key) error) error {
	return db.ForEachPrefixKeyOnly(keyPrefix, func(key database.Key) (stop bool) {
		return consume(key) != nil
	})
}

func (db *BoltDB) batchOperation(batchOpcount int32, operation func(opIndex int32, batch objectstorage.BatchedMutations)) error {
	batchAmount := int(math.Ceil(float64(batchOpcount) / float64(BoltDBSubBatchSize)))
	for i := 0; i < batchAmount; i++ {
		batch := db.bolt.Batched()
		batchStart := int32(i * BoltDBSubBatchSize)
		batchEnd := batchStart + int32(BoltDBSubBatchSize)

		if batchEnd > batchOpcount {
			batchEnd = batchOpcount
		}

		for j := batchStart; j < batchEnd; j++ {
			operation(j, batch)
		}

		//fmt.Printf("Applied %d/%d insertions\n", batchEnd, batchOpcount)
		err := batch.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}

// Transactions
func (db *BoltDB) Apply(set []database.Entry, delete []database.Key) error {

	setCount := int32(len(set))
	deleteCount := int32(len(delete))

	err := db.batchOperation(setCount, func(i int32, batch objectstorage.BatchedMutations) {
		batch.Set(set[i].Key, set[i].Value)
	})
	if err != nil {
		return err
	}
	err = db.batchOperation(deleteCount, func(i int32, batch objectstorage.BatchedMutations) {
		batch.Delete(delete[i])
	})

	return err
}
