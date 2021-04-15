package events

import (
	"strconv"
	"sync"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/datastructure/thresholdmap"
	"github.com/iotaledger/hive.go/marshalutil"
)

// region ThresholdEvent ///////////////////////////////////////////////////////////////////////////////////////////////

type ThresholdEvent struct {
	*Event
	thresholds         *thresholdmap.ThresholdMap
	currentLevels      map[interface{}]int
	currentLevelsMutex sync.Mutex
}

func NewThresholdEvent(eventHandler ThresholdHandler, thresholds ...float64) (thresholdEvent *ThresholdEvent) {
	thresholdEvent = &ThresholdEvent{
		Event: NewEvent(func(handler interface{}, params ...interface{}) {
			eventHandler(handler, params[0].(interface{}), params[1].(int), params[2].(ThresholdLevelTransition))
		}),
		currentLevels: make(map[interface{}]int),
		thresholds:    thresholdmap.New(thresholdmap.LowerThresholdMode),
	}

	for i, threshold := range thresholds {
		thresholdEvent.registerThreshold(threshold, i+1)
	}

	return
}

func ThresholdEventFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil, eventHandler ThresholdHandler, identifierParser func(marshalUtil *marshalutil.MarshalUtil) (identifier interface{}, err error), thresholds ...float64) (thresholdEvent *ThresholdEvent, consumedBytes int, err error) {
	thresholdEvent = &ThresholdEvent{
		Event: NewEvent(func(handler interface{}, params ...interface{}) {
			eventHandler(handler, params[0].(interface{}), params[1].(int), params[2].(ThresholdLevelTransition))
		}),
		currentLevels: make(map[interface{}]int),
		thresholds:    thresholdmap.New(thresholdmap.LowerThresholdMode),
	}

	for i, threshold := range thresholds {
		thresholdEvent.registerThreshold(threshold, i+1)
	}

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

		identifier, identifierErr := identifierParser(marshalUtil)
		if identifierErr != nil {
			err = xerrors.Errorf("failed to read identifier (%v): %w", identifierErr, cerrors.ErrParseBytesFailed)
			return
		}

		thresholdEvent.currentLevels[identifier] = int(value)
	}

	return
}

func (t *ThresholdEvent) RegisterThreshold(threshold float64) (level int) {
	level = t.thresholds.Size()
	t.registerThreshold(threshold, level)

	return
}

func (t *ThresholdEvent) Trigger(identifier marshalutil.SimpleBinaryMarshaler, value float64) {
	t.currentLevelsMutex.Lock()
	defer t.currentLevelsMutex.Unlock()

	newLevel, levelReached := t.level(value)
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
			t.Event.Trigger(branchID, i, ThresholdLevelIncreased)
		}
	} else {
		for i := oldLevel - 1; i >= newLevel; i-- {
			t.Event.Trigger(branchID, i, ThresholdLevelDecreased)
		}
	}
}

func (t *ThresholdEvent) registerThreshold(threshold float64, level int) {
	t.thresholds.Set(threshold, level)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ThresholdLevelTransition /////////////////////////////////////////////////////////////////////////////////////

type ThresholdLevelTransition int

const (
	ThresholdLevelIncreased ThresholdLevelTransition = 1
	ThresholdLevelDecreased ThresholdLevelTransition = -1
)

func (t ThresholdLevelTransition) String() string {
	switch t {
	case 1:
		return "ThresholdLevelIncreased"
	case -1:
		return "ThresholdLevelDecreased"
	default:
		return "ThresholdLevelTransition(" + strconv.Itoa(int(t)) + ")"
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ThresholdHandler /////////////////////////////////////////////////////////////////////////////////////////////

type ThresholdHandler func(handler interface{}, identifier interface{}, newLevel int, transition ThresholdLevelTransition)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
