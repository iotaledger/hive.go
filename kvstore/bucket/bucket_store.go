package bucketstore

import (
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/atomic"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/kvstore"
)

var (
	// ErrNotEnoughBuckets is returned when not enough buckets are available.
	ErrNotEnoughBuckets = errors.New("not enough buckets")
	// ErrInvalidBucketIndex is returned when a bucket index does not exist or is invalid.
	ErrInvalidBucketIndex = errors.New("invalid bucket index")
)

const (
	// MinBucketCount is the minimum amount of remaining buckets in the store.
	MinBucketCount = 2
)

// BucketCaller is an event caller which gets a bucket passed.
func BucketCaller(handler interface{}, params ...interface{}) {
	handler.(func(kvstore.KVStore))(params[0].(kvstore.KVStore))
}

// BucketConsumerFunc consumes the given index and bucket.
// A returned boolean flag signals to proceed further iteration.
type BucketConsumerFunc = func(index int, bucket kvstore.KVStore) bool

// Events are the events issued by the BucketStore.
type Events struct {
	// PreBucketCreated is fired before a new bucket is created.
	PreBucketCreated *events.Event
	// PreBucketDeleted is fired before a bucket is deleted.
	PreBucketDeleted *events.Event
	// BucketCreated is fired when a new bucket is created.
	BucketCreated *events.Event
	// BucketDeleted is fired when a bucket is deleted.
	BucketDeleted *events.Event
}

// BucketStore is a KVStore implementation that consists of multiple "buckets"
// of other KVStores.
//
// Operations applied to all buckets:
//    - Get
//    - Has
//    - Iterate (in order from latest bucket to oldest bucket)
//    - IterateKeys (in order from latest bucket to oldest bucket)
//    - Delete
//    - DeletePrefix
//    - Clear
//    - Close
//    - Flush
//
// Operations applied to the latest bucket:
//    - Set
//
// Important to consider
//
// During a Set operation, old versions of existing keys
// in older buckets are deleted, if WithDeleteKeyInOtherBucketsOnSet
// is set to true (default).
//
// If this option is set to false, there could be multiple versions
// of the same key.
//
// Get always returns the latest version.
// Iterate and IterateKeys will return the different versions
// of the keys in order from latest bucket to oldest bucket.
//
// Setting this option to true may impact the write performance.
type BucketStore struct {
	sync.RWMutex

	// holds the buckets.
	buckets []kvstore.KVStore
	// holds the BucketStore options.
	opts *Options
	// events of the BucketStore.
	Events *Events
	// used to signal that the store was closed already.
	closed *atomic.Bool
}

// the default options applied to the BucketStore.
var defaultOptions = []Option{
	WithDeleteKeyInOtherBucketsOnSet(true),
}

// Options define options for the BucketStore.
type Options struct {
	// deletes older versions of the key in other buckets on Set or Batched().Set
	deleteKeyInOtherBucketsOnSet bool
}

// applies the given Option.
func (so *Options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(so)
	}
}

// WithDeleteKeyInOtherBucketsOnSet deletes older versions of the key in other buckets on Set or Batched().Set.
func WithDeleteKeyInOtherBucketsOnSet(deleteOnSet bool) Option {
	return func(opts *Options) {
		opts.deleteKeyInOtherBucketsOnSet = deleteOnSet
	}
}

// Option is a function setting a BucketStore option.
type Option func(opts *Options)

// New creates a new BucketStore.
func New(buckets []kvstore.KVStore, opts ...Option) (*BucketStore, error) {
	if len(buckets) < MinBucketCount {
		return nil, ErrNotEnoughBuckets
	}

	options := &Options{}
	options.apply(defaultOptions...)
	options.apply(opts...)

	return &BucketStore{
		buckets: buckets,
		opts:    options,
		closed:  atomic.NewBool(false),

		Events: &Events{
			PreBucketCreated: events.NewEvent(BucketCaller),
			PreBucketDeleted: events.NewEvent(events.VoidCaller),
			BucketCreated:    events.NewEvent(BucketCaller),
			BucketDeleted:    events.NewEvent(events.VoidCaller),
		},
	}, nil
}

// Options returns a copy of the options.
func (bs *BucketStore) Options() Options {
	return *bs.opts
}

// bucketForEachWithoutLocking calls the consumer function on every bucket.
func (bs *BucketStore) bucketForEachWithoutLocking(consumer BucketConsumerFunc) bool {

	for i, bucket := range bs.buckets {
		if !consumer(i, bucket) {
			return false
		}
	}

	return true
}

// BucketAdd adds a new bucket to the store (front).
func (bs *BucketStore) BucketAdd(bucket kvstore.KVStore) error {
	if bs.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	bs.Events.PreBucketCreated.Trigger(bucket)

	// push front
	bs.Lock()
	bs.buckets = append([]kvstore.KVStore{bucket}, bs.buckets...)
	bs.Unlock()

	bs.Events.BucketCreated.Trigger(bucket)

	return nil
}

// BucketDeleteLast deletes a bucket from the store (end).
func (bs *BucketStore) BucketDeleteLast() error {
	if bs.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	bs.RLock()
	lenBuckets := len(bs.buckets)
	bs.RUnlock()

	if lenBuckets <= MinBucketCount {
		return ErrNotEnoughBuckets
	}

	bs.Events.PreBucketDeleted.Trigger()

	bs.Lock()

	// check again
	lenBuckets = len(bs.buckets)
	if lenBuckets <= MinBucketCount {
		bs.Unlock()
		return ErrNotEnoughBuckets
	}

	lastBucket := bs.buckets[lenBuckets-1]
	lastBucket.Close()
	bs.buckets[lenBuckets-1] = nil // avoid memory leak

	// pop
	bs.buckets = bs.buckets[:lenBuckets-1]
	bs.Unlock()

	bs.Events.BucketDeleted.Trigger()
	return nil
}

func (bs *BucketStore) bucketFirst() kvstore.KVStore {
	return bs.buckets[0]
}

// BucketFirst returns the latest bucket (front).
func (bs *BucketStore) BucketFirst() (kvstore.KVStore, error) {
	if bs.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	return bs.bucketFirst(), nil
}

func (bs *BucketStore) bucketLast() kvstore.KVStore {
	return bs.buckets[len(bs.buckets)-1]
}

// BucketLast returns the oldest bucket (end).
func (bs *BucketStore) BucketLast() (kvstore.KVStore, error) {
	if bs.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	return bs.bucketLast(), nil
}

// Bucket returns the bucket at the given index.
func (bs *BucketStore) Bucket(index int) (kvstore.KVStore, error) {
	if bs.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	if index < 0 || index >= len(bs.buckets) {
		return nil, ErrInvalidBucketIndex
	}

	return bs.buckets[index], nil
}

// BucketCount returns the amount of buckets in the store.
func (bs *BucketStore) BucketCount() (int, error) {
	if bs.closed.Load() {
		return 0, kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	return len(bs.buckets), nil
}

// BucketForEach calls the consumer on every bucket in order from latest to oldest.
func (bs *BucketStore) BucketForEach(consumer BucketConsumerFunc) (bool, error) {
	if bs.closed.Load() {
		return false, kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	return bs.bucketForEachWithoutLocking(consumer), nil
}

///////////////////////
// KVStore interface //
///////////////////////

// WithRealm is a factory method for using the same underlying storage with a different realm.
func (bs *BucketStore) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	if bs.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	return NewRealm(bs, realm), nil
}

// Realm returns the configured realm.
func (bs *BucketStore) Realm() kvstore.Realm {
	bs.RLock()
	defer bs.RUnlock()

	return bs.bucketFirst().Realm()
}

// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (bs *BucketStore) Iterate(prefix kvstore.KeyPrefix, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc, direction ...kvstore.IterDirection) error {
	if bs.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	var innerErr error

	proceedIteration := true

	kvConsumerFuncWrapped := func(key kvstore.Key, value kvstore.Value) bool {
		if !kvConsumerFunc(key, value) {
			proceedIteration = false
		}
		return proceedIteration
	}

	bs.bucketForEachWithoutLocking(func(index int, bucket kvstore.KVStore) bool {
		if err := bucket.Iterate(prefix, kvConsumerFuncWrapped, direction...); err != nil {
			innerErr = err
			return false
		}
		return proceedIteration
	})

	return innerErr
}

// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (bs *BucketStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, direction ...kvstore.IterDirection) error {
	if bs.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	var innerErr error

	proceedIteration := true

	consumerFuncWrapped := func(key kvstore.Key) bool {
		if !consumerFunc(key) {
			proceedIteration = false
		}
		return proceedIteration
	}

	bs.bucketForEachWithoutLocking(func(index int, bucket kvstore.KVStore) bool {
		if err := bucket.IterateKeys(prefix, consumerFuncWrapped, direction...); err != nil {
			innerErr = err
			return false
		}
		return proceedIteration
	})

	return innerErr
}

// Clear clears the realm.
func (bs *BucketStore) Clear() error {
	if bs.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	var innerErr error

	bs.bucketForEachWithoutLocking(func(index int, bucket kvstore.KVStore) bool {
		if err := bucket.Clear(); err != nil {
			innerErr = err
		}

		// clear all buckets, even if one fails.
		return true
	})

	return innerErr
}

// Get gets the given key or nil if it doesn't exist or an error if an error occurred.
func (bs *BucketStore) Get(key kvstore.Key) (kvstore.Value, error) {
	if bs.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	var value kvstore.Value = nil
	var innerErr error = kvstore.ErrKeyNotFound

	bs.bucketForEachWithoutLocking(func(index int, bucket kvstore.KVStore) bool {
		v, err := bucket.Get(key)
		if err != nil {
			if err == kvstore.ErrKeyNotFound {
				// key not found.
				// => continue searching for the key in other buckets.
				return true
			}

			// an error has occurred.
			// => stop searching for the key in other buckets.
			innerErr = err
			return false
		}

		// reset the ErrKeyNotFound because we found the key.
		innerErr = nil
		value = v

		// key found.
		// => stop searching for the key in other buckets.
		return false
	})

	return value, innerErr
}

// Set sets the given key and value.
func (bs *BucketStore) Set(key kvstore.Key, value kvstore.Value) error {
	if bs.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	if bs.opts.deleteKeyInOtherBucketsOnSet {
		_ = bs.bucketForEachWithoutLocking(func(index int, bucket kvstore.KVStore) bool {
			// Delete the key in older buckets
			if index != 0 {
				_ = bucket.Delete(key)
			}
			return true
		})
	}

	return bs.bucketFirst().Set(key, value)
}

// Has checks whether the given key exists.
func (bs *BucketStore) Has(key kvstore.Key) (bool, error) {
	if bs.closed.Load() {
		return false, kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	var has bool = false
	var innerErr error = nil

	bs.bucketForEachWithoutLocking(func(index int, bucket kvstore.KVStore) bool {
		found, err := bucket.Has(key)
		if err != nil {
			// an error has occurred.
			// => stop searching for the key in other buckets.
			innerErr = err
			return false
		}

		if found {
			has = true

			// key found.
			// => stop searching for the key in other buckets.
			return false
		}

		// key not found.
		// => continue searching for the key in other buckets.
		return true
	})

	return has, innerErr
}

// Delete deletes the entry for the given key.
func (bs *BucketStore) Delete(key kvstore.Key) error {
	if bs.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	var innerErr error

	bs.bucketForEachWithoutLocking(func(index int, bucket kvstore.KVStore) bool {
		if err := bucket.Delete(key); err != nil {
			innerErr = err
		}

		// delete key in all buckets, even if one fails.
		return true
	})

	return innerErr
}

// DeletePrefix deletes all the entries matching the given key prefix.
func (bs *BucketStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	if bs.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	var innerErr error

	bs.bucketForEachWithoutLocking(func(index int, bucket kvstore.KVStore) bool {
		if err := bucket.DeletePrefix(prefix); err != nil {
			innerErr = err
		}

		// delete prefix in all buckets, even if one fails.
		return true
	})

	return innerErr
}

// Flush persists all outstanding write operations to disc.
func (bs *BucketStore) Flush() error {
	if bs.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	var innerErr error

	bs.bucketForEachWithoutLocking(func(index int, bucket kvstore.KVStore) bool {
		if err := bucket.Flush(); err != nil {
			innerErr = err
		}

		// flush all buckets, even if one fails.
		return true
	})

	return innerErr
}

// Close closes the database file handles.
func (bs *BucketStore) Close() error {
	if bs.closed.Swap(true) {
		// was already closed
		return kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	var innerErr error

	bs.bucketForEachWithoutLocking(func(index int, bucket kvstore.KVStore) bool {
		if err := bucket.Close(); err != nil {
			innerErr = err
		}

		// close all buckets, even if one fails.
		return true
	})

	return innerErr
}

// Batched returns a BatchedMutations interface to execute batched mutations.
func (bs *BucketStore) Batched() (kvstore.BatchedMutations, error) {
	if bs.closed.Load() {
		return nil, kvstore.ErrStoreClosed
	}

	bs.RLock()
	defer bs.RUnlock()

	batchedOthers := make([]kvstore.BatchedMutations, len(bs.buckets)-1)

	var innerErr error

	_ = bs.bucketForEachWithoutLocking(func(index int, bucket kvstore.KVStore) bool {
		// Initialize a batched mutation for the other buckets,
		// because we need to delete all keys that are deleted, and also delete
		// keys that are set if "deleteKeyInOtherBucketsOnSet" is set to true.
		if index != 0 {
			batchedMutation, err := bucket.Batched()
			if err != nil {
				innerErr = err
				return false
			}
			batchedOthers[index-1] = batchedMutation
		}
		return true
	})

	if innerErr != nil {
		return nil, innerErr
	}

	batchedFirst, err := bs.bucketFirst().Batched()
	if err != nil {
		return nil, err
	}

	return &batchedMutations{
		batchedFirst:                 batchedFirst,
		batchedOthers:                batchedOthers,
		deleteKeyInOtherBucketsOnSet: bs.opts.deleteKeyInOtherBucketsOnSet,
		closed:                       bs.closed,
	}, nil
}

// batchedMutations is a wrapper around a kvstore.BatchedMutations.
type batchedMutations struct {
	batchedFirst                 kvstore.BatchedMutations
	batchedOthers                []kvstore.BatchedMutations
	deleteKeyInOtherBucketsOnSet bool
	closed                       *atomic.Bool
}

// Set sets the given key and value.
func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {

	if b.deleteKeyInOtherBucketsOnSet {
		for _, batched := range b.batchedOthers {
			if err := batched.Delete(key); err != nil {
				return err
			}
		}
	}

	return b.batchedFirst.Set(key, value)
}

// Delete deletes the entry for the given key.
func (b *batchedMutations) Delete(key kvstore.Key) error {
	for _, batched := range b.batchedOthers {
		if err := batched.Delete(key); err != nil {
			return err
		}
	}

	return b.batchedFirst.Delete(key)
}

// Cancel cancels the batched mutations.
func (b *batchedMutations) Cancel() {
	for _, batched := range b.batchedOthers {
		batched.Cancel()
	}

	b.batchedFirst.Cancel()
}

// Commit commits/flushes the mutations.
func (b *batchedMutations) Commit() error {
	if b.closed.Load() {
		return kvstore.ErrStoreClosed
	}

	for _, batched := range b.batchedOthers {
		if err := batched.Commit(); err != nil {
			return err
		}
	}

	return b.batchedFirst.Commit()
}

// code guards
var _ kvstore.KVStore = &BucketStore{}
var _ kvstore.BatchedMutations = &batchedMutations{}
