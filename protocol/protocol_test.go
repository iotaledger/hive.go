package protocol_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"
	"testing"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/protocol"
	"github.com/iotaledger/hive.go/protocol/message"
	"github.com/iotaledger/hive.go/protocol/tlv"
	"github.com/stretchr/testify/assert"
)

type fakeconn struct {
	writer io.WriteCloser
	reader io.ReadCloser
}

func (f fakeconn) Read(p []byte) (n int, err error) {
	return f.reader.Read(p)
}

func (f fakeconn) Write(p []byte) (n int, err error) {
	return f.writer.Write(p)
}

func (f fakeconn) Close() error {
	if err := f.writer.Close(); err != nil {
		return err
	}
	if err := f.reader.Close(); err != nil {
		return err
	}
	return nil
}

func newFakeConn() *fakeconn {
	r, w := io.Pipe()
	return &fakeconn{writer: w, reader: r}
}

func consume(t *testing.T, p *protocol.Protocol, conn io.Reader, expectedLength int) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		data := make([]byte, expectedLength)
		read, err := conn.Read(data)
		assert.NoError(t, err)
		assert.Equal(t, expectedLength, read)
		p.Read(data[:read])
	}()
	return &wg
}

const (
	MessageTypeTest message.Type = 1

	// length of a test message in bytes
	TestMaxBytesLength = 5
)

var (
	TestMessageDefinition = &message.Definition{
		ID:             MessageTypeTest,
		MaxBytesLength: TestMaxBytesLength,
		VariableLength: false,
	}
	msgRegistry = message.NewRegistry([]*message.Definition{
		tlv.HeaderMessageDefinition,
		TestMessageDefinition,
	})
)

func newTestMessage() ([]byte, error) {
	packet := []byte{'t', 'e', 's', 't', '!'}
	// create a buffer for tlv header plus the packet
	buf := bytes.NewBuffer(make([]byte, 0, tlv.HeaderMessageDefinition.MaxBytesLength+uint16(TestMaxBytesLength)))
	// write tlv header into buffer
	if err := tlv.WriteHeader(buf, MessageTypeTest, uint16(TestMaxBytesLength)); err != nil {
		return nil, err
	}
	// write serialized packet bytes into the buffer
	if err := binary.Write(buf, binary.BigEndian, packet); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func TestMessageReceive(t *testing.T) {
	conn := newFakeConn()
	defer conn.Close()
	p := protocol.New(msgRegistry)

	var TestMessageReceived bool
	var testPacketString string
	p.Events.Received[TestMessageDefinition.ID].Attach(events.NewClosure(func(data []byte) {
		TestMessageReceived = true
		testPacketString = string(data)
	}))

	testMsg, err := newTestMessage()
	assert.NoError(t, err)

	wg := consume(t, p, conn, len(testMsg))
	_, err = conn.Write(testMsg)
	assert.NoError(t, err)

	wg.Wait()
	assert.True(t, TestMessageReceived)
	assert.Equal(t, "test!", testPacketString)
}
