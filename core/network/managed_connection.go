package network

import (
	"net"
	"sync"
	"time"

	"go.uber.org/atomic"
)

// ManagedConnection provides a wrapper for a net.Conn to be used together with events.
type ManagedConnection struct {
	net.Conn
	Events *ManagedConnectionEvents

	readTimeout  time.Duration
	writeTimeout time.Duration
	closeOnce    sync.Once

	bytesRead    atomic.Uint64
	bytesWritten atomic.Uint64
}

func NewManagedConnection(conn net.Conn) *ManagedConnection {
	return &ManagedConnection{
		Conn:   conn,
		Events: newManagedConnectionEvents(),
	}
}

// BytesRead returns the total number of bytes read.
func (mc *ManagedConnection) BytesRead() uint64 {
	return mc.bytesRead.Load()
}

// BytesWritten returns the total number of bytes written.
func (mc *ManagedConnection) BytesWritten() uint64 {
	return mc.bytesWritten.Load()
}

func (mc *ManagedConnection) Read(p []byte) (int, error) {
	read := 0
	for {
		if err := mc.setReadTimeoutBasedDeadline(); err != nil {
			return read, err
		}

		n, err := mc.Conn.Read(p)
		read += n
		mc.bytesRead.Add(uint64(n))
		if err != nil {
			mc.Events.Error.Trigger(err)

			return read, err
		}
		if n > 0 {
			// copy the data before triggering
			receivedData := make([]byte, n)
			copy(receivedData, p)
			mc.Events.ReceiveData.Trigger(&ReceivedDataEvent{receivedData})
		}
	}
}

func (mc *ManagedConnection) Write(p []byte) (int, error) {
	if err := mc.setWriteTimeoutBasedDeadline(); err != nil {
		return 0, err
	}

	n, err := mc.Conn.Write(p)
	mc.bytesWritten.Add(uint64(n))

	return n, err
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (mc *ManagedConnection) Close() (err error) {
	mc.closeOnce.Do(func() {
		// do not trigger the error event to prevent deadlocks
		err = mc.Conn.Close()
		// trigger Close event in separate go routine to prevent deadlocks
		go mc.Events.Close.Trigger(&CloseEvent{})
	})

	return err
}

func (mc *ManagedConnection) SetTimeout(d time.Duration) error {
	if err := mc.SetReadTimeout(d); err != nil {
		return err
	}

	if err := mc.SetWriteTimeout(d); err != nil {
		return err
	}

	return nil
}

func (mc *ManagedConnection) SetReadTimeout(d time.Duration) error {
	mc.readTimeout = d

	if err := mc.setReadTimeoutBasedDeadline(); err != nil {
		return err
	}

	return nil
}

func (mc *ManagedConnection) SetWriteTimeout(d time.Duration) error {
	mc.writeTimeout = d

	if err := mc.setWriteTimeoutBasedDeadline(); err != nil {
		return err
	}

	return nil
}

func (mc *ManagedConnection) setReadTimeoutBasedDeadline() error {
	if mc.readTimeout != 0 {
		if err := mc.Conn.SetReadDeadline(time.Now().Add(mc.readTimeout)); err != nil {
			return err
		}
	} else {
		if err := mc.Conn.SetReadDeadline(time.Time{}); err != nil {
			return err
		}
	}

	return nil
}

func (mc *ManagedConnection) setWriteTimeoutBasedDeadline() error {
	if mc.writeTimeout != 0 {
		if err := mc.Conn.SetWriteDeadline(time.Now().Add(mc.writeTimeout)); err != nil {
			return err
		}
	} else {
		if err := mc.Conn.SetWriteDeadline(time.Time{}); err != nil {
			return err
		}
	}

	return nil
}
