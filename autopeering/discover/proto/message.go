package proto

import (
	"google.golang.org/protobuf/proto"

	"github.com/izuc/zipp.foundation/autopeering/server"
)

// MType is the type of message type enum.
type MType = server.MType

// An enum for the different message types.
const (
	MPing MType = 10 + iota
	MPong
	MDiscoveryRequest
	MDiscoveryResponse
)

// Message extends the proto.Message interface with additional util functions.
type Message interface {
	proto.Message

	// Name returns the name of the corresponding message type for debugging.
	Name() string
	// Type returns the type of the corresponding message as an enum.
	Type() MType
}

func (x *Ping) Name() string { return "PING" }
func (x *Ping) Type() MType  { return MPing }

func (x *Pong) Name() string { return "PONG" }
func (x *Pong) Type() MType  { return MPong }

func (x *DiscoveryRequest) Name() string { return "DISCOVERY_REQUEST" }
func (x *DiscoveryRequest) Type() MType  { return MDiscoveryRequest }

func (x *DiscoveryResponse) Name() string { return "DISCOVERY_RESPONSE" }
func (x *DiscoveryResponse) Type() MType  { return MDiscoveryResponse }
