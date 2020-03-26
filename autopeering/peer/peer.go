package peer

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
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
	ErrInvalidSignature    = errors.New("invalid signature")
)

// PublicKey is the type of Ed25519 public keys used for peers.
type PublicKey ed25519.PublicKey

// Peer defines the immutable data of a peer.
type Peer struct {
	*identity.Identity
	ip       net.IP
	services *service.Record // unmodifiable services supported by the peer
}

// IP returns the public IP of the peer.
func (p *Peer) IP() net.IP {
	return p.ip
}

// Network returns the autopeering network of the peer.
func (p *Peer) Network() string {
	return p.services.Get(service.PeeringKey).Network()
}

// Address returns the autopeering address of a peer.
func (p *Peer) Address() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   p.ip,
		Port: p.services.Get(service.PeeringKey).Port(),
	}
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
		Host:   p.Address().String(),
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
func NewPeer(id *identity.Identity, ip net.IP, services service.Service) *Peer {
	if services.Get(service.PeeringKey) == nil {
		panic("need peering service")
	}

	return &Peer{
		Identity: id,
		ip:       ip,
		services: services.CreateRecord(),
	}
}

// ToProto encodes a given peer into a proto buffer Peer message
func (p *Peer) ToProto() *pb.Peer {
	return &pb.Peer{
		PublicKey: p.PublicKey().Bytes(),
		Ip:        p.IP().String(),
		Services:  p.services.ToProto(),
	}
}

// FromProto decodes a given proto buffer Peer message (in) and returns the corresponding Peer.
func FromProto(in *pb.Peer) (*Peer, error) {
	publicKey, err, _ := ed25519.PublicKeyFromBytes(in.GetPublicKey())
	if err != nil {
		return nil, err
	}
	id := identity.NewIdentity(publicKey)

	ip := net.ParseIP(in.GetIp())
	if ip == nil {
		return nil, fmt.Errorf("invalid IP: %s", in.GetIp())
	}

	services, err := service.FromProto(in.GetServices())
	if err != nil {
		return nil, err
	}
	if services.Get(service.PeeringKey) == nil {
		return nil, ErrNeedsPeeringService
	}

	return NewPeer(id, ip, services), nil
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
