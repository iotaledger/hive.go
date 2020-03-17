package server

import (
	"crypto/sha256"
	"net"

	"github.com/golang/protobuf/proto"
	"github.com/iotaledger/hive.go/autopeering/peer"
)

const (
	// MaxPacketSize specifies the maximum allowed size of packets.
	// Packets larger than this will be cut and thus treated as invalid.
	MaxPacketSize = 1280
)

// MType is the type of message type enum.
type MType uint32

// Message extends the proto.testMessage interface with additional type.
type Message interface {
	proto.Message

	// Type returns the type of the corresponding message as an enum.
	Type() MType
}

// The Sender interface specifies common method required to send requests.
type Sender interface {
	Send(toAddr *net.UDPAddr, data []byte)
	SendExpectingReply(toAddr *net.UDPAddr, toID peer.ID, data []byte, replyType MType, callback func(interface{}) bool) <-chan error
}

// A Handler reacts to an incoming message.
type Handler interface {
	// HandleMessage is called for each incoming message.
	// It returns true, if that particular message type can be processed by the current Handler.
	HandleMessage(s *Server, fromAddr *net.UDPAddr, fromID peer.ID, fromKey peer.PublicKey, data []byte) (bool, error)
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as Server handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler that calls f.
type HandlerFunc func(*Server, *net.UDPAddr, peer.ID, peer.PublicKey, []byte) (bool, error)

// HandleMessage returns f(s, from, data).
func (f HandlerFunc) HandleMessage(s *Server, fromAddr *net.UDPAddr, fromID peer.ID, fromKey peer.PublicKey, data []byte) (bool, error) {
	return f(s, fromAddr, fromID, fromKey, data)
}

// PacketHash returns the hash of a packet
func PacketHash(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}
