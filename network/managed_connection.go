package network

import (
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/events"
	"go.uber.org/atomic"
)

type ManagedConnection struct {
	Conn         net.Conn
	Events       BufferedConnectionEvents
	readTimeout  time.Duration
	writeTimeout time.Duration
	closeOnce    sync.Once

	bytesRead    atomic.Uint64
	bytesWritten atomic.Uint64
}

func NewManagedConnection(conn net.Conn) *ManagedConnection {
	bufferedConnection := &ManagedConnection{
		Conn: conn,
		Events: BufferedConnectionEvents{
			ReceiveData: events.NewEvent(events.ByteSliceCaller),
			Close:       events.NewEvent(events.CallbackCaller),
			Error:       events.NewEvent(events.ErrorCaller),
		},
	}

	return bufferedConnection
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
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic while reading from socket", r, string(debug.Stack()))
		}
		mc.Close()
	}()

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
			mc.Events.ReceiveData.Trigger(receivedData)
		}
	}
}

func (mc *ManagedConnection) Write(p []byte) (int, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic while writing to socket", r)
			mc.Close()
		}
	}()
	if err := mc.setWriteTimeoutBasedDeadline(); err != nil {
		return 0, err
	}

	n, err := mc.Conn.Write(p)
	mc.bytesWritten.Add(uint64(n))
	return n, err
}

func (mc *ManagedConnection) Close() error {
	err := mc.Conn.Close()
	if err != nil {
		mc.Events.Error.Trigger(err)
	}

	mc.closeOnce.Do(func() {
		go mc.Events.Close.Trigger()
	})

	return err
}

func (mc *ManagedConnection) LocalAddr() net.Addr {
	return mc.Conn.LocalAddr()
}

func (mc *ManagedConnection) RemoteAddr() net.Addr {
	return mc.Conn.RemoteAddr()
}

func (mc *ManagedConnection) SetDeadline(t time.Time) error {
	return mc.Conn.SetDeadline(t)
}

func (mc *ManagedConnection) SetReadDeadline(t time.Time) error {
	return mc.Conn.SetReadDeadline(t)
}

func (mc *ManagedConnection) SetWriteDeadline(t time.Time) error {
	return mc.Conn.SetWriteDeadline(t)
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
