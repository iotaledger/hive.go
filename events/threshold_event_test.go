package events

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/cerrors"
	"github.com/iotaledger/hive.go/marshalutil"
)

func identifierTypeThresholdEventCaller(handler interface{}, identifier interface{}, newLevel int, transition ThresholdLevelTransition) {
	handler.(func(id identifierType, newLevel int, transition ThresholdLevelTransition))(identifier.(identifierType), newLevel, transition)
}

func TestThresholdEvent(t *testing.T) {
	thresholds := []float64{0.2, 0.4, 0.8}

	thresholdEvent := NewThresholdEvent(identifierTypeThresholdEventCaller, thresholds...)

	thresholdEvent.Attach(NewClosure(func(identifier identifierType, newLevel int, transition ThresholdLevelTransition) {
		fmt.Println(identifier, newLevel, transition)
	}))

	thresholdEvent.Trigger(identifierType(1), 0.2)
	thresholdEvent.Trigger(identifierType(1), 0.21)
	thresholdEvent.Trigger(identifierType(1), 0.41)
	thresholdEvent.Trigger(identifierType(1), 0.38)
	thresholdEvent.Trigger(identifierType(1), 0.19)
	thresholdEvent.Trigger(identifierType(1), 0.90)

	unmarshaledEvent, _, err := ThresholdEventFromMarshalUtil(marshalutil.New(thresholdEvent.Bytes()), identifierTypeThresholdEventCaller, identifierTypeFromMarshalUtil, thresholds...)
	assert.NoError(t, err)

	unmarshaledEvent.Attach(NewClosure(func(identifier identifierType, newLevel int, transition ThresholdLevelTransition) {
		fmt.Println(identifier, newLevel, transition)
	}))

	unmarshaledEvent.Trigger(identifierType(1), 0.1)
}

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
