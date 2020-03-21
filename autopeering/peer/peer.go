package peer

import (
	"encoding/base64"
	"errors"
	"net/url"

	"github.com/golang/protobuf/proto"
	pb "github.com/iotaledger/hive.go/autopeering/peer/proto"
	"github.com/iotaledger/hive.go/autopeering/peer/service"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/identity"
)

// Errors in the peer package.
var (
	ErrNeedsPeeringService = errors.New("needs peering service")
)

// Peer defines the immutable data of a peer.
type Peer struct {
	*identity.Identity
	services *service.Record // unmodifiable services supported by the peer
}

// Network returns the autopeering network of the peer.
func (p *Peer) Network() string {
	return p.services.Get(service.PeeringKey).Network()
}

// Address returns the autopeering address of a peer.
func (p *Peer) Address() string {
	return p.services.Get(service.PeeringKey).String()
}

// Services returns the supported services of the peer.
func (p *Peer) Services() service.Service {
	return p.services
}

// String returns a string representation of the peer.
func (p *Peer) String() string {
	u := url.URL{
		Scheme: "peer",
		User:   url.User(base64.StdEncoding.EncodeToString(p.PublicKey().Bytes())),
		Host:   p.Address(),
	}
	return u.String()
}

// SignedData is an interface wrapper around data with key and signature.
type SignedData interface {
	GetData() []byte
	GetPublicKey() []byte
	GetSignature() []byte
}

// RecoverKeyFromSignedData validates and returns the key that was used to sign the data.
func RecoverKeyFromSignedData(m SignedData) (ed25519.PublicKey, error) {
	return ed25519.RecoverKey(m.GetPublicKey(), m.GetData(), m.GetSignature())
}

// NewPeer creates a new unmodifiable peer.
func NewPeer(publicKey ed25519.PublicKey, services service.Service) *Peer {
	if services.Get(service.PeeringKey) == nil {
		panic("need peering service")
	}

	return &Peer{
		Identity: identity.NewIdentity(publicKey),
		services: services.CreateRecord(),
	}
}

func NewPeerWithIdentity(identity *identity.Identity, services service.Service) *Peer {
	if services.Get(service.PeeringKey) == nil {
		panic("need peering service")
	}

	return &Peer{
		Identity: identity,
		services: services.CreateRecord(),
	}
}

// ToProto encodes a given peer into a proto buffer Peer message
func (p *Peer) ToProto() *pb.Peer {
	return &pb.Peer{
		PublicKey: p.PublicKey().Bytes(),
		Services:  p.services.ToProto(),
	}
}

// FromProto decodes a given proto buffer Peer message (in) and returns the corresponding Peer.
func FromProto(in *pb.Peer) (*Peer, error) {
	publicKey, err, _ := ed25519.PublicKeyFromBytes(in.GetPublicKey())
	if err != nil {
		return nil, err
	}

	services, err := service.FromProto(in.GetServices())
	if err != nil {
		return nil, err
	}
	if services.Get(service.PeeringKey) == nil {
		return nil, ErrNeedsPeeringService
	}

	return NewPeer(publicKey, services), nil
}

// Marshal serializes a given Peer (p) into a slice of bytes.
func (p *Peer) Marshal() ([]byte, error) {
	return proto.Marshal(p.ToProto())
}

// Unmarshal de-serializes a given slice of bytes (data) into a Peer.
func Unmarshal(data []byte) (*Peer, error) {
	s := &pb.Peer{}
	if err := proto.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return FromProto(s)
}
