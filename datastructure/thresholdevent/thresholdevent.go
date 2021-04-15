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
// using the Set method causes the identifier to reach a new threshold.
type ThresholdEvent struct {
	thresholds         *thresholdmap.ThresholdMap
	currentLevels      map[interface{}]int
	currentLevelsMutex sync.Mutex
	configuration      *Configuration

	*events.Event
	objectstorage.StorableObjectFlags
}

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

func FromMarshalUtil(marshalUtil *marshalutil.MarshalUtil, options ...Option) (thresholdEvent *ThresholdEvent, consumedBytes int, err error) {
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

func (t *ThresholdEvent) Set(identifier Identifier, newValue float64) {
	t.currentLevelsMutex.Lock()
	defer t.currentLevelsMutex.Unlock()

	newLevel, levelReached := t.level(newValue)
	if !levelReached {
		if currentLevel, exists := t.currentLevels[identifier]; exists {
			delete(t.currentLevels, identifier)

			t.trigger(identifier, currentLevel, newLevel)
		}

		return
	}

	currentLevel := t.currentLevels[identifier]
	if currentLevel == newLevel {
		return
	}

	t.currentLevels[identifier] = newLevel

	t.trigger(identifier, currentLevel, newLevel)
}

func (t *ThresholdEvent) Bytes() []byte {
	return byteutils.ConcatBytes(t.ObjectStorageKey(), t.ObjectStorageValue())
}

func (t *ThresholdEvent) Update(objectstorage.StorableObject) {
	panic("updates disabled")
}

func (t *ThresholdEvent) ObjectStorageKey() []byte {
	return t.configuration.ObjectStorageKey
}

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

func (t *ThresholdEvent) level(value float64) (level int, levelReached bool) {
	untypedLevel, exists := t.thresholds.Get(value)
	if !exists {
		return 0, false
	}

	return untypedLevel.(int), true
}

func (t *ThresholdEvent) trigger(branchID interface{}, oldLevel, newLevel int) {
	if newLevel >= oldLevel {
		for i := oldLevel + 1; i <= newLevel; i++ {
			t.Event.Trigger(branchID, i, LevelIncreased)
		}
	} else {
		for i := oldLevel - 1; i >= newLevel; i-- {
			t.Event.Trigger(branchID, i, LevelDecreased)
		}
	}
}

func (t *ThresholdEvent) registerThreshold(threshold float64, level int) {
	t.thresholds.Set(threshold, level)
}

var _ objectstorage.StorableObject = &ThresholdEvent{}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Configuration ////////////////////////////////////////////////////////////////////////////////////////////////

type Configuration struct {
	Thresholds         []float64
	CallbackTypecaster CallbackTypecaster
	IdentifierParser   IdentifierParser
	ObjectStorageKey   []byte
}

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

type Option func(*Configuration)

func WithObjectStorageKey(key []byte) Option {
	return func(configuration *Configuration) {
		configuration.ObjectStorageKey = key
	}
}

func WithThresholds(thresholds ...float64) Option {
	return func(options *Configuration) {
		options.Thresholds = thresholds
	}
}

func WithIdentifierParser(identifierParser IdentifierParser) Option {
	return func(configuration *Configuration) {
		configuration.IdentifierParser = identifierParser
	}
}

func WithCallbackTypeCaster(callbackTypeCaster CallbackTypecaster) Option {
	return func(configuration *Configuration) {
		configuration.CallbackTypecaster = callbackTypeCaster
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region LevelTransition //////////////////////////////////////////////////////////////////////////////////////////////

type LevelTransition int

const (
	LevelIncreased LevelTransition = 1
	LevelDecreased LevelTransition = -1
)

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

type Identifier marshalutil.SimpleBinaryMarshaler

type CallbackTypecaster func(handler interface{}, identifier interface{}, newLevel int, transition LevelTransition)

type IdentifierParser func(marshalUtil *marshalutil.MarshalUtil) (identifier interface{}, err error)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
