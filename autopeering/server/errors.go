package server

import "github.com/iotaledger/hive.go/ierrors"

var (
	// ErrTimeout is returned when an expected response was not received in time.
	ErrTimeout = ierrors.New("response timeout")

	// ErrClosed means that the server was shut down before a response could be received.
	ErrClosed = ierrors.New("socket closed")

	// ErrNoMessage is returned when the package did not contain any data.
	ErrNoMessage = ierrors.New("packet does not contain a message")

	// ErrInvalidMessage means that no handler could process the received message.
	ErrInvalidMessage = ierrors.New("invalid message")
)
