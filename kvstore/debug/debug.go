package debug

import (
	"github.com/iotaledger/hive.go/ds/bitmask"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

// Command is a type that represents a specific method in the KVStore.
type Command = bitmask.BitMask

// AccessCallback is the type of the callback function that can be used to hook the access to the callback.
type AccessCallback func(command Command, parameters ...[]byte)

const (
	// IterateCommand represents a call to the Iterate method of the store.
	IterateCommand Command = 1 << iota

	// IterateKeysCommand represents a call to the IterateKeys method of the store.
	IterateKeysCommand

	// ClearCommand represents a call to the Clear method of the store.
	ClearCommand

	// GetCommand represents a call to the Get method of the store.
	GetCommand

	// SetCommand represents a call to the Set method of the store.
	SetCommand

	// HasCommand represents a call to the Has method of the store.
	HasCommand

	// DeleteCommand represents a call to the Delete method of the store.
	DeleteCommand

	// DeletePrefixCommand represents a call to the DeletePrefix method of the store.
	DeletePrefixCommand

	// AllCommands represents the collection of all commands.
	AllCommands = IterateCommand | IterateKeysCommand | ClearCommand | GetCommand | SetCommand | HasCommand | DeleteCommand | DeletePrefixCommand

	// ShutdownCommand represents a call to the Shutdown method of the store.
	ShutdownCommand Command = 0
)

// CommandNames contains a map from the command to its human-readable name.
var CommandNames = map[Command]string{
	ShutdownCommand:     "Shutdown",
	IterateCommand:      "Iterate",
	IterateKeysCommand:  "IterateKeys",
	ClearCommand:        "Clear",
	GetCommand:          "Get",
	SetCommand:          "Set",
	HasCommand:          "Has",
	DeleteCommand:       "Delete",
	DeletePrefixCommand: "DeletePrefix",
}

// DebugStore implements the KVStore interface wrapping another KVStore to provide access callbacks for debug purposes.
type debugStore struct {
	underlying                   kvstore.KVStore
	accessCallback               AccessCallback
	accessCallbackCommandsFilter Command
}

// New creates a new KVStore with debug callbacks. This can for example be used for debugging and to examine what the KVStore is doing.
// store is the underlying store to which all methods will be forwarded.
// callback will be called for all requests to the KVStore.
// commandsFilter optional filter to only report certain requests to the callback.
func New(store kvstore.KVStore, callback AccessCallback, commandsFilter ...Command) kvstore.KVStore {
	var accessCallbackCommandsFilter Command
	if len(commandsFilter) == 0 {
		accessCallbackCommandsFilter = AllCommands
	} else {
		for _, filterCommand := range commandsFilter {
			accessCallbackCommandsFilter |= filterCommand
		}
	}

	return &debugStore{
		underlying:                   store,
		accessCallback:               callback,
		accessCallbackCommandsFilter: accessCallbackCommandsFilter,
	}
}

func (s *debugStore) WithRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	storeWithRealm, err := s.underlying.WithRealm(realm)
	if err != nil {
		return nil, err
	}

	return &debugStore{
		underlying:                   storeWithRealm,
		accessCallback:               s.accessCallback,
		accessCallbackCommandsFilter: s.accessCallbackCommandsFilter,
	}, nil
}

func (s *debugStore) WithExtendedRealm(realm kvstore.Realm) (kvstore.KVStore, error) {
	return s.WithRealm(byteutils.ConcatBytes(s.Realm(), realm))
}

func (s *debugStore) Realm() kvstore.Realm {
	return s.underlying.Realm()
}

// Iterate iterates over all keys and values with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys and values.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *debugStore) Iterate(prefix kvstore.KeyPrefix, kvConsumerFunc kvstore.IteratorKeyValueConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(IterateCommand) {
		s.accessCallback(IterateCommand, prefix)
	}

	return s.underlying.Iterate(prefix, kvConsumerFunc, iterDirection...)
}

// IterateKeys iterates over all keys with the provided prefix. You can pass kvstore.EmptyPrefix to iterate over all keys.
// Optionally the direction for the iteration can be passed (default: IterDirectionForward).
func (s *debugStore) IterateKeys(prefix kvstore.KeyPrefix, consumerFunc kvstore.IteratorKeyConsumerFunc, iterDirection ...kvstore.IterDirection) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(IterateKeysCommand) {
		s.accessCallback(IterateKeysCommand, prefix)
	}

	return s.underlying.IterateKeys(prefix, consumerFunc, iterDirection...)
}

func (s *debugStore) Clear() error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(ClearCommand) {
		s.accessCallback(ClearCommand)
	}

	return s.underlying.Clear()
}

func (s *debugStore) Get(key kvstore.Key) (value kvstore.Value, err error) {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(GetCommand) {
		s.accessCallback(GetCommand, key)
	}

	return s.underlying.Get(key)
}

func (s *debugStore) Set(key kvstore.Key, value kvstore.Value) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(SetCommand) {
		s.accessCallback(SetCommand, key, value)
	}

	return s.underlying.Set(key, value)
}

func (s *debugStore) Has(key kvstore.Key) (bool, error) {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(HasCommand) {
		s.accessCallback(HasCommand, key)
	}

	return s.underlying.Has(key)
}

func (s *debugStore) Delete(key kvstore.Key) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(DeleteCommand) {
		s.accessCallback(DeleteCommand, key)
	}

	return s.underlying.Delete(key)
}

func (s *debugStore) DeletePrefix(prefix kvstore.KeyPrefix) error {
	if s.accessCallback != nil && s.accessCallbackCommandsFilter.HasBits(DeletePrefixCommand) {
		s.accessCallback(DeletePrefixCommand, prefix)
	}

	return s.underlying.DeletePrefix(prefix)
}

func (s *debugStore) Flush() error {
	return s.underlying.Flush()
}

func (s *debugStore) Close() error {
	return s.underlying.Close()
}

func (s *debugStore) Batched() (kvstore.BatchedMutations, error) {
	batchedMutation, err := s.underlying.Batched()
	if err != nil {
		return nil, err
	}

	return &batchedMutations{
		underlying:                   batchedMutation,
		accessCallback:               s.accessCallback,
		accessCallbackCommandsFilter: s.accessCallbackCommandsFilter,
	}, nil
}

type batchedMutations struct {
	underlying                   kvstore.BatchedMutations
	accessCallback               AccessCallback
	accessCallbackCommandsFilter Command
}

func (b *batchedMutations) Set(key kvstore.Key, value kvstore.Value) error {
	if b.accessCallback != nil && b.accessCallbackCommandsFilter.HasBits(SetCommand) {
		b.accessCallback(SetCommand, key, value)
	}

	return b.underlying.Set(key, value)
}

func (b *batchedMutations) Delete(key kvstore.Key) error {
	if b.accessCallback != nil && b.accessCallbackCommandsFilter.HasBits(DeleteCommand) {
		b.accessCallback(DeleteCommand, key)
	}

	return b.underlying.Delete(key)
}

func (b *batchedMutations) Cancel() {
	b.underlying.Cancel()
}

func (b *batchedMutations) Commit() error {
	return b.underlying.Commit()
}

var _ kvstore.KVStore = &debugStore{}
var _ kvstore.BatchedMutations = &batchedMutations{}
