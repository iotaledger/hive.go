package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/atomic"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/cerrors"
	"github.com/iotaledger/hive.go/core/marshalutil"
)

// region TestThresholdEvent ///////////////////////////////////////////////////////////////////////////////////////////

func TestThresholdEvent(t *testing.T) {
	options := []ThresholdEventOption{
		WithThresholds(0.2, 0.4, 0.8),
		WithCallbackTypeCaster(func(handler interface{}, identifier interface{}, newLevel int, transition ThresholdEventTransition) {
			handler.(func(id identifierType, newLevel int, transition ThresholdEventTransition))(identifier.(identifierType), newLevel, transition)
		}),
		WithIdentifierParser(identifierTypeFromMarshalUtil),
	}

	eventHandler := newMockedEventHandler(t)

	thresholdEvent := NewThresholdEvent(options...)
	thresholdEvent.Attach(NewClosure(eventHandler.Trigger))

	eventHandler.Expect(identifierType(1), 1, ThresholdLevelIncreased)
	thresholdEvent.Set(identifierType(1), 0.2)
	eventHandler.AssertExpectations()

	thresholdEvent.Set(identifierType(1), 0.21)
	eventHandler.AssertExpectations()

	eventHandler.Expect(identifierType(1), 2, ThresholdLevelIncreased)
	thresholdEvent.Set(identifierType(1), 0.41)
	eventHandler.AssertExpectations()

	eventHandler.Expect(identifierType(1), 1, ThresholdLevelDecreased)
	thresholdEvent.Set(identifierType(1), 0.38)
	eventHandler.AssertExpectations()

	eventHandler.Expect(identifierType(1), 0, ThresholdLevelDecreased)
	thresholdEvent.Set(identifierType(1), 0.19)
	eventHandler.AssertExpectations()

	eventHandler.Expect(identifierType(1), 1, ThresholdLevelIncreased)
	thresholdEvent.Set(identifierType(1), 0.38)
	eventHandler.AssertExpectations()

	eventHandler.Expect(identifierType(1), 2, ThresholdLevelIncreased)
	eventHandler.Expect(identifierType(1), 3, ThresholdLevelIncreased)
	thresholdEvent.Set(identifierType(1), 0.90)
	eventHandler.AssertExpectations()

	unmarshaledEvent, err := ThresholdEventFromMarshalUtil(marshalutil.New(thresholdEvent.Bytes()), options...)
	assert.NoError(t, err)
	unmarshaledEvent.Attach(NewClosure(eventHandler.Trigger))

	eventHandler.Expect(identifierType(1), 2, ThresholdLevelDecreased)
	eventHandler.Expect(identifierType(1), 1, ThresholdLevelDecreased)
	eventHandler.Expect(identifierType(1), 0, ThresholdLevelDecreased)
	unmarshaledEvent.Set(identifierType(1), 0.1)
	eventHandler.AssertExpectations()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region mockedEventHandler ///////////////////////////////////////////////////////////////////////////////////////////

type mockedEventHandler struct {
	mock.Mock
	expectedTriggers *atomic.Uint64
	observedTriggers *atomic.Uint64
	test             *testing.T
}

func newMockedEventHandler(t *testing.T) *mockedEventHandler {
	result := &mockedEventHandler{
		expectedTriggers: atomic.NewUint64(0),
		observedTriggers: atomic.NewUint64(0),
		test:             t,
	}
	result.Test(t)

	return result
}

func (e *mockedEventHandler) Trigger(identifier identifierType, newLevel int, transition ThresholdEventTransition) {
	e.Called(identifier, newLevel, transition)
	e.observedTriggers.Inc()
}

func (e *mockedEventHandler) Expect(arguments ...interface{}) {
	e.On("Trigger", arguments...)
	e.expectedTriggers.Inc()
}

func (e *mockedEventHandler) AssertExpectations() bool {
	calledEvents := e.observedTriggers.Load()
	expectedEvents := e.expectedTriggers.Load()
	if calledEvents != expectedEvents {
		e.test.Errorf("number of called (%d) events is not equal to number of expected events (%d)", calledEvents, expectedEvents)
		return false
	}

	defer func() {
		e.Calls = make([]mock.Call, 0)
		e.ExpectedCalls = make([]*mock.Call, 0)
		e.expectedTriggers.Store(0)
		e.observedTriggers.Store(0)
	}()

	return e.Mock.AssertExpectations(e.test)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region identifierType ///////////////////////////////////////////////////////////////////////////////////////////////

type identifierType uint64

func identifierTypeFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (identifier interface{}, err error) {
	untypedIdentifierType, err := marshalUtil.ReadUint64()
	if err != nil {
		err = xerrors.Errorf("Failed to read identifier type (%v): %w", err, cerrors.ErrParseBytesFailed)
	}
	identifier = identifierType(untypedIdentifierType)

	return
}

func (i identifierType) Bytes() []byte {
	return marshalutil.New(marshalutil.Uint64Size).
		WriteUint64(uint64(i)).
		Bytes()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
