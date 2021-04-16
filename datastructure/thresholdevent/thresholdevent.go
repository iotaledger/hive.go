package thresholdevent

import (
	"fmt"
	"sync"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/datastructure/thresholdmap"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/iotaledger/hive.go/objectstorage"
)

// region ThresholdEvent ///////////////////////////////////////////////////////////////////////////////////////////////

// ThresholdEvent is a data structure that acts like a normal event but only triggers when the value that was reported
// using the Set method causes the corresponding identifier to reach a new threshold. It is stateful which means that it
// tracks the current level for all identifiers individually to not trigger the same event multiple times.
type ThresholdEvent struct {
	thresholds         *thresholdmap.ThresholdMap
	currentLevels      map[interface{}]int
	currentLevelsMutex sync.Mutex
	configuration      *Configuration

	*events.Event
	objectstorage.StorableObjectFlags
}

// New is the constructor for the ThresholdEvent.
func New(options ...Option) (thresholdEvent *ThresholdEvent) {
	thresholdEvent = &ThresholdEvent{
		currentLevels: make(map[interface{}]int),
		thresholds:    thresholdmap.New(thresholdmap.LowerThresholdMode),
		configuration: NewConfiguration(options...),
	}

	if thresholdEvent.configuration.Thresholds == nil {
		panic("missing WithThresholds option when calling New")
	}

	for i, threshold := range thresholdEvent.configuration.Thresholds {
		thresholdEvent.registerThreshold(threshold, i+1)
	}

	thresholdEvent.Event = events.NewEvent(func(handler interface{}, params ...interface{}) {
		thresholdEvent.configuration.CallbackTypecaster(handler, params[0].(interface{}), params[1].(int), params[2].(LevelTransition))
	})

	return
}

// FromBytes unmarshals a collection of BranchIDs from a sequence of bytes.
func FromBytes(bytes []byte, options ...Option) (thresholdEvent *ThresholdEvent, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if thresholdEvent, err = FromMarshalUtil(marshalUtil, options...); err != nil {
		err = xerrors.Errorf("failed to parse ThresholdEvent from MarshalUtil: %w", err)
		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// FromMarshalUtil unmarshals a ThresholdEvent using a MarshalUtil (for easier unmarshaling).
func FromMarshalUtil(marshalUtil *marshalutil.MarshalUtil, options ...Option) (thresholdEvent *ThresholdEvent, err error) {
	thresholdEvent = &ThresholdEvent{
		currentLevels: make(map[interface{}]int),
		thresholds:    thresholdmap.New(thresholdmap.LowerThresholdMode),
		configuration: NewConfiguration(options...),
	}

	if thresholdEvent.configuration.Thresholds == nil {
		panic("missing WithThresholds option when calling FromMarshalUtil")
	}
	if thresholdEvent.configuration.IdentifierParser == nil {
		panic("missing WithIdentifierParser option when calling FromMarshalUtil")
	}

	for i, threshold := range thresholdEvent.configuration.Thresholds {
		thresholdEvent.registerThreshold(threshold, i+1)
	}

	thresholdEvent.Event = events.NewEvent(func(handler interface{}, params ...interface{}) {
		thresholdEvent.configuration.CallbackTypecaster(handler, params[0].(interface{}), params[1].(int), params[2].(LevelTransition))
	})

	levelCount, err := marshalUtil.ReadUint64()
	if err != nil {
		err = xerrors.Errorf("failed to read level count (%v): %w", err, cerrors.ErrParseBytesFailed)
		return
	}

	for i := uint64(0); i < levelCount; i++ {
		value, valueErr := marshalUtil.ReadInt64()
		if valueErr != nil {
			err = xerrors.Errorf("failed to read level (%v): %w", valueErr, cerrors.ErrParseBytesFailed)
			return
		}

		identifier, identifierErr := thresholdEvent.configuration.IdentifierParser(marshalUtil)
		if identifierErr != nil {
			err = xerrors.Errorf("failed to read identifier (%v): %w", identifierErr, cerrors.ErrParseBytesFailed)
			return
		}

		thresholdEvent.currentLevels[identifier] = int(value)
	}

	return
}

// FromObjectStorage restores a ThresholdEvent that was stored in the object storage.
func FromObjectStorage(key []byte, data []byte) (branch objectstorage.StorableObject, err error) {
	if branch, _, err = FromBytes(byteutils.ConcatBytes(key, data)); err != nil {
		err = xerrors.Errorf("failed to parse ThresholdEvent from bytes: %w", err)
		return
	}

	return
}

// Set updates the value associated with the given identifier and triggers the Event if necessary.
func (t *ThresholdEvent) Set(identifier Identifier, newValue float64) (newLevel int, transition LevelTransition) {
	t.currentLevelsMutex.Lock()
	defer t.currentLevelsMutex.Unlock()

	newLevel, levelReached := t.level(newValue)
	if !levelReached {
		if currentLevel, exists := t.currentLevels[identifier]; exists {
			delete(t.currentLevels, identifier)

			transition = t.trigger(identifier, currentLevel, newLevel)
		}

		return
	}

	currentLevel := t.currentLevels[identifier]
	if currentLevel == newLevel {
		return
	}

	t.currentLevels[identifier] = newLevel

	transition = t.trigger(identifier, currentLevel, newLevel)

	return
}

// Level returns the current level of the reached threshold for the given identity.
func (t *ThresholdEvent) Level(identifier Identifier) (level int) {
	t.currentLevelsMutex.Lock()
	defer t.currentLevelsMutex.Unlock()

	return t.currentLevels[identifier]
}

// Bytes returns a marshaled version of the ThresholdEvent.
func (t *ThresholdEvent) Bytes() []byte {
	return byteutils.ConcatBytes(t.ObjectStorageKey(), t.ObjectStorageValue())
}

// Update is disabled and panics if it ever gets called - it is required to match the StorableObject interface.
func (t *ThresholdEvent) Update(objectstorage.StorableObject) {
	panic("updates disabled")
}

// ObjectStorageKey returns the key that is used to store the object in the database. It is required to match the
// StorableObject interface.
func (t *ThresholdEvent) ObjectStorageKey() []byte {
	return t.configuration.ObjectStorageKey
}

// ObjectStorageValue marshals the ConflictBranch into a sequence of bytes that are used as the value part in the
// object storage.
func (t *ThresholdEvent) ObjectStorageValue() []byte {
	t.currentLevelsMutex.Lock()
	defer t.currentLevelsMutex.Unlock()

	marshalUtil := marshalutil.New()
	marshalUtil.WriteUint64(uint64(len(t.currentLevels)))
	for key, value := range t.currentLevels {
		marshalUtil.WriteInt64(int64(value))
		marshalUtil.WriteBytes(key.(marshalutil.SimpleBinaryMarshaler).Bytes())
	}

	return marshalUtil.Bytes()
}

// level returns the level of the threshold that the given value represents (and a boolean flag indicating if no
// threshold was reached).
func (t *ThresholdEvent) level(value float64) (level int, levelReached bool) {
	untypedLevel, exists := t.thresholds.Get(value)
	if !exists {
		return 0, false
	}

	return untypedLevel.(int), true
}

// trigger triggers the embedded Event with the correct parameters.
func (t *ThresholdEvent) trigger(branchID interface{}, oldLevel, newLevel int) (transition LevelTransition) {
	if newLevel >= oldLevel {
		transition = LevelIncreased

		for i := oldLevel + 1; i <= newLevel; i++ {
			t.Event.Trigger(branchID, i, transition)
		}
	} else {
		transition = LevelDecreased

		for i := oldLevel - 1; i >= newLevel; i-- {
			t.Event.Trigger(branchID, i, transition)
		}
	}

	return
}

// registerThreshold create a new threshold in the internal ThresholdMap.
func (t *ThresholdEvent) registerThreshold(threshold float64, level int) {
	t.thresholds.Set(threshold, level)
}

// code contract (make sure the struct implements all required methods)
var _ objectstorage.StorableObject = &ThresholdEvent{}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Configuration ////////////////////////////////////////////////////////////////////////////////////////////////

// Configuration represents a collection of optional parameters that are used by the ThresholdEvent.
type Configuration struct {
	Thresholds         []float64
	CallbackTypecaster CallbackTypecaster
	IdentifierParser   IdentifierParser
	ObjectStorageKey   []byte
}

// NewConfiguration creates a Configuration from the given Options.
func NewConfiguration(options ...Option) (configuration *Configuration) {
	configuration = &Configuration{
		Thresholds: make([]float64, 0),
		CallbackTypecaster: func(handler interface{}, identifier interface{}, newLevel int, transition LevelTransition) {
			handler.(func(identifier interface{}, newLevel int, transition LevelTransition))(identifier, newLevel, transition)
		},
	}
	for _, option := range options {
		option(configuration)
	}

	return configuration
}

// Option is the type of the optional parameters of the ThresholdEvent.
type Option func(*Configuration)

// WithObjectStorageKey sets the object storage key that is used for persisting the ThresholdEvent.
func WithObjectStorageKey(key []byte) Option {
	return func(configuration *Configuration) {
		configuration.ObjectStorageKey = key
	}
}

// WithThresholds sets the thresholds that are supposed to be used for the Triggers.
func WithThresholds(thresholds ...float64) Option {
	return func(options *Configuration) {
		options.Thresholds = thresholds
	}
}

// WithIdentifierParser sets the parser for the Identifier that is used to identify different entities.
func WithIdentifierParser(identifierParser IdentifierParser) Option {
	return func(configuration *Configuration) {
		configuration.IdentifierParser = identifierParser
	}
}

// WithCallbackTypeCaster sets the method that is used to type cast the called callbacks to their correct types.
func WithCallbackTypeCaster(callbackTypeCaster CallbackTypecaster) Option {
	return func(configuration *Configuration) {
		configuration.CallbackTypecaster = callbackTypeCaster
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region LevelTransition //////////////////////////////////////////////////////////////////////////////////////////////

// LevelTransition is the type of the values that are used to indicate in which direction a threshold was passed.
type LevelTransition int

const (
	// LevelMaintained indicates that the reached threshold did not change.
	LevelMaintained LevelTransition = 0

	// LevelIncreased indicates that the new value is larger than the passed threshold.
	LevelIncreased LevelTransition = 1

	// LevelDecreased indicates that the new value is smaller than the passed threshold.
	LevelDecreased LevelTransition = -1
)

// String returns a human readable version of the LevelTransition.
func (t LevelTransition) String() string {
	switch t {
	case 1:
		return "LevelIncreased"
	case -1:
		return "LevelDecreased"
	default:
		panic(fmt.Sprintf("invalid LevelTransition (%d)", int(t)))
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Types & Interfaces ///////////////////////////////////////////////////////////////////////////////////////////

// Identifier is the type that is used to address the identifiers of the entities whose values we are tracking.
type Identifier marshalutil.SimpleBinaryMarshaler

// CallbackTypecaster defines the signature of the function that is used to convert the parameters to the types expected
// by the callbacks.
type CallbackTypecaster func(handler interface{}, identifier interface{}, newLevel int, transition LevelTransition)

// IdentifierParser defines the signature of the function that is used to parse the Identifiers.
type IdentifierParser func(marshalUtil *marshalutil.MarshalUtil) (identifier interface{}, err error)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
