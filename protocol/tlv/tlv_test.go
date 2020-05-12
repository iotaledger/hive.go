package tlv_test

import (
	"bytes"
	"testing"

	"github.com/iotaledger/hive.go/protocol/message"
	"github.com/iotaledger/hive.go/protocol/tlv"
	"github.com/stretchr/testify/assert"
)

func init() {
	r = message.NewRegistry()
	if err := r.RegisterType(MessageTypeTest, TestMessageDefinition); err != nil {
		panic(err)
	}
}

const(
	MessageTypeTest message.Type = 1

	// length of a test message in bytes
	TestMaxBytesLength = 5
)

var (
	r *message.Registry
	TestMessageDefinition = &message.Definition{
		ID:             MessageTypeTest,
		MaxBytesLength: TestMaxBytesLength,
		VariableLength: false,
	}
)

func TestTLV(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, tlv.HeaderMessageDefinition.MaxBytesLength))

	err :=tlv.WriteHeader(buf, MessageTypeTest, uint16(TestMaxBytesLength))
	assert.NoError(t, err)

	data := buf.Bytes()
	var header *tlv.Header
	header, err = tlv.ParseHeader(data, r)
	assert.NoError(t, err)

	assert.Equal(t, TestMessageDefinition, header.Definition)
	assert.Equal(t, uint16(TestMaxBytesLength), header.MessageBytesLength)
}

func TestTLV_ParseHeader(t *testing.T) {
	// unknown message type
	data := []byte{2,0,5}
	_, err := tlv.ParseHeader(data, r)
	assert.Error(t, err)

	// invalid message length
	data = []byte{byte(MessageTypeTest),0,6}
	_, err = tlv.ParseHeader(data, r)
	assert.Error(t, err)

}