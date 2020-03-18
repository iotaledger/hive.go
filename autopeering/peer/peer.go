package peer

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"

	"github.com/golang/protobuf/proto"
	pb "github.com/iotaledger/hive.go/autopeering/peer/proto"
	"github.com/iotaledger/hive.go/autopeering/peer/service"
	"github.com/iotaledger/hive.go/signature"
)

// Errors in the peer package.
var (
	ErrNeedsPeeringService = errors.New("needs peering service")
	ErrInvalidSignature    = errors.New("invalid signature")
)

// Peer defines the immutable data of a peer.
type Peer struct {
	id        ID                  // comparable node identifier
	publicKey signature.PublicKey // public key used to verify signatures
	services  *service.Record     // unmodifiable services supported by the peer
}

// ID returns the identifier of the peer.
func (p *Peer) ID() ID {
	return p.id
}

// PublicKey returns the public key of the peer.
func (p *Peer) PublicKey() signature.PublicKey {
	return p.publicKey
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
		User:   url.User(base64.StdEncoding.EncodeToString(p.PublicKey())),
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
func RecoverKeyFromSignedData(m SignedData) (signature.PublicKey, error) {
	return recoverKey(m.GetPublicKey(), m.GetData(), m.GetSignature())
}

// NewPeer creates a new unmodifiable peer.
func NewPeer(publicKey signature.PublicKey, services service.Service) *Peer {
	if services.Get(service.PeeringKey) == nil {
		panic("need peering service")
	}

	return &Peer{
		id:        CreateID(publicKey),
		publicKey: publicKey,
		services:  services.CreateRecord(),
	}
}

// ToProto encodes a given peer into a proto buffer Peer message
func (p *Peer) ToProto() *pb.Peer {
	return &pb.Peer{
		PublicKey: p.publicKey,
		Services:  p.services.ToProto(),
	}
}

// FromProto decodes a given proto buffer Peer message (in) and returns the corresponding Peer.
func FromProto(in *pb.Peer) (*Peer, error) {
	if l := len(in.GetPublicKey()); l != signature.PublicKeySize {
		return nil, fmt.Errorf("invalid key length: %d, need %d", l, signature.PublicKeySize)
	}
	services, err := service.FromProto(in.GetServices())
	if err != nil {
		return nil, err
	}
	if services.Get(service.PeeringKey) == nil {
		return nil, ErrNeedsPeeringService
	}

	return NewPeer(in.GetPublicKey(), services), nil
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

func recoverKey(key, data, sig []byte) (signature.PublicKey, error) {
	if l := len(key); l != signature.PublicKeySize {
		return nil, fmt.Errorf("%w: invalid key length: %d, need %d", ErrInvalidSignature, l, signature.PublicKeySize)
	}
	if l := len(sig); l != signature.SignatureSize {
		return nil, fmt.Errorf("%w: invalid signature length: %d, need %d", ErrInvalidSignature, l, signature.SignatureSize)
	}
	if !signature.Verify(key, data, sig) {
		return nil, ErrInvalidSignature
	}

	publicKey := make([]byte, signature.PublicKeySize)
	copy(publicKey, key)
	return publicKey, nil
}
