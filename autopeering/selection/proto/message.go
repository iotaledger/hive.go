package proto

import (
	"google.golang.org/protobuf/proto"

	"github.com/izuc/zipp.foundation/autopeering/server"
)

// MType is the type of message type enum.
type MType = server.MType

// An enum for the different message types.
const (
	MPeeringRequest MType = 20 + iota
	MPeeringResponse
	MPeeringDrop
)

// Message extends the proto.Message interface with additional util functions.
type Message interface {
	proto.Message

	// Name returns the name of the corresponding message type for debugging.
	Name() string
	// Type returns the type of the corresponding message as an enum.
	Type() MType
}

func (x *PeeringRequest) Name() string { return "PEERING_REQUEST" }
func (x *PeeringRequest) Type() MType  { return MPeeringRequest }

func (x *PeeringResponse) Name() string { return "PEERING_RESPONSE" }
func (x *PeeringResponse) Type() MType  { return MPeeringResponse }

func (x *PeeringDrop) Name() string { return "PEERING_DROP" }
func (x *PeeringDrop) Type() MType  { return MPeeringDrop }
