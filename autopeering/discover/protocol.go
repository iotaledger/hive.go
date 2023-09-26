package discover

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/izuc/zipp.foundation/autopeering/discover/proto"
	"github.com/izuc/zipp.foundation/autopeering/netutil"
	"github.com/izuc/zipp.foundation/autopeering/peer"
	peerpb "github.com/izuc/zipp.foundation/autopeering/peer/proto"
	"github.com/izuc/zipp.foundation/autopeering/peer/service"
	"github.com/izuc/zipp.foundation/autopeering/server"
	"github.com/izuc/zipp.foundation/crypto/identity"
	"github.com/izuc/zipp.foundation/logger"
	"github.com/izuc/zipp.foundation/runtime/backoff"
)

const (
	backoffInterval = 500 * time.Millisecond
	maxRetries      = 2
	logSends        = true
)

// policy for retrying failed network calls.
var retryPolicy = backoff.ExponentialBackOff(backoffInterval, 1.5).With(
	backoff.Jitter(0.5), backoff.MaxRetries(maxRetries))

// The Protocol handles the peer discovery.
// It responds to incoming messages and sends own requests when needed.
type Protocol struct {
	server.Protocol

	loc     *peer.Local    // local peer that runs the protocol
	version uint32         // version number of the protocol
	netID   uint32         // network ID of the protocol
	log     *logger.Logger // protocol logger

	mgr       *manager // the manager handles the actual peer discovery and re-verification
	running   atomic.Bool
	closeOnce sync.Once
}

// New creates a new discovery protocol for the local node with the given protocol version and networkID.
func New(local *peer.Local, version uint32, networkID uint32, opts ...Option) *Protocol {
	args := &options{
		log:         logger.NewNopLogger(),
		masterPeers: nil,
	}
	for _, opt := range opts {
		opt.apply(args)
	}

	p := &Protocol{
		loc:     local,
		version: version,
		netID:   networkID,
		log:     args.log,
	}

	p.mgr = newManager(p, args.masterPeers, args.log.Named("mgr"))

	return p
}

// Start starts the actual peer discovery over the provided Sender.
func (p *Protocol) Start(s server.Sender) {
	p.Protocol.Sender = s
	p.mgr.start()
	p.log.Debug("discover started")
	p.running.Store(true)
}

// Close finalizes the protocol.
func (p *Protocol) Close() {
	p.closeOnce.Do(func() {
		p.running.Store(false)
		p.mgr.close()
	})
}

// Events returns all the events that are triggered during the peer discovery.
func (p *Protocol) Events() *Events {
	return p.mgr.events
}

// IsVerified checks whether the given peer has recently been verified a recent enough endpoint proof.
func (p *Protocol) IsVerified(id identity.ID, ip net.IP) bool {
	return time.Since(p.loc.Database().LastPong(id, ip)) < PingExpiration
}

// EnsureVerified checks if the given peer has recently sent a Ping;
// if not, we send a Ping to trigger a verification.
func (p *Protocol) EnsureVerified(peer *peer.Peer) error {
	if !p.hasVerified(peer.ID(), peer.IP()) {
		if err := p.Ping(peer); err != nil {
			return err
		}
		// Wait for them to Ping back and process our pong
		time.Sleep(server.ResponseTimeout)
	}

	return nil
}

// GetMasterPeers returns the list of master peers.
func (p *Protocol) GetMasterPeers() []*peer.Peer {
	return unwrapPeers(p.mgr.masterPeers())
}

// GetVerifiedPeer returns the verified peer with the given ID, or nil if no such peer exists.
func (p *Protocol) GetVerifiedPeer(id identity.ID) *peer.Peer {
	for _, verified := range p.mgr.verifiedPeers() {
		if verified.ID() == id {
			return unwrapPeer(verified)
		}
	}
	// if the sender is not managed, try to load it from DB
	from, err := p.local().Database().Peer(id)
	if err != nil {
		// this should not happen as this is checked in validation
		p.log.Warnw("invalid stored peer",
			"id", id,
			"err", err,
		)

		return nil
	}
	// send ping to restored peer to ensure that it will be verified
	p.sendPing(from.Address(), from.ID())

	return from
}

// GetVerifiedPeers returns all the currently managed peers that have been verified at least once.
func (p *Protocol) GetVerifiedPeers() []*peer.Peer {
	return unwrapPeers(p.mgr.verifiedPeers())
}

// HandleMessage responds to incoming peer discovery messages.
func (p *Protocol) HandleMessage(s *server.Server, fromAddr *net.UDPAddr, from *identity.Identity, data []byte) (bool, error) {
	if !p.running.Load() {
		return false, nil
	}

	switch pb.MType(data[0]) {
	// Ping
	case pb.MPing:
		m := new(pb.Ping)
		if err := proto.Unmarshal(data[1:], m); err != nil {
			return true, fmt.Errorf("invalid message: %w", err)
		}
		if p.validatePing(fromAddr, m) {
			p.handlePing(s, fromAddr, from, m, data)
		}

	// Pong
	case pb.MPong:
		m := new(pb.Pong)
		if err := proto.Unmarshal(data[1:], m); err != nil {
			return true, fmt.Errorf("invalid message: %w", err)
		}
		if p.validatePong(s, fromAddr, from.ID(), m) {
			p.handlePong(fromAddr, from, m)
		}

	// DiscoveryRequest
	case pb.MDiscoveryRequest:
		m := new(pb.DiscoveryRequest)
		if err := proto.Unmarshal(data[1:], m); err != nil {
			return true, fmt.Errorf("invalid message: %w", err)
		}
		if p.validateDiscoveryRequest(fromAddr, from.ID(), m) {
			p.handleDiscoveryRequest(s, from.ID(), data)
		}

	// DiscoveryResponse
	case pb.MDiscoveryResponse:
		m := new(pb.DiscoveryResponse)
		if err := proto.Unmarshal(data[1:], m); err != nil {
			return true, fmt.Errorf("invalid message: %w", err)
		}
		p.validateDiscoveryResponse(s, fromAddr, from.ID(), m)
		// DiscoveryResponse messages are handled in the handleReply function of the validation

	default:
		return false, nil
	}

	return true, nil
}

// local returns the associated local peer of the neighbor selection.
func (p *Protocol) local() *peer.Local {
	return p.loc
}

// localAddr returns the address under which the peering service can be reached.
func (p *Protocol) localAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   p.loc.IP(),
		Port: p.loc.Services().Get(service.PeeringKey).Port(),
	}
}

// ------ message senders ------

// Ping sends a Ping to the specified peer and blocks until a reply is received or timeout.
func (p *Protocol) Ping(to *peer.Peer) error {
	return backoff.Retry(retryPolicy, func() error {
		err := <-p.sendPing(to.Address(), to.ID())
		if err != nil && !errors.Is(err, server.ErrTimeout) {
			return backoff.Permanent(err)
		}

		return err
	})
}

// sendPing sends a Ping to the specified address and expects a matching reply.
// This method is non-blocking, but it returns a channel that can be used to query potential errors.
func (p *Protocol) sendPing(toAddr *net.UDPAddr, toID identity.ID) <-chan error {
	// set the src address to zero to force response to the UDP envelop address
	srcAddr := p.localAddr()
	srcAddr.IP = net.IPv4zero

	ping := newPing(p.version, p.netID, srcAddr, toAddr)
	data := marshal(ping)

	// compute the message hash
	hash := server.PacketHash(data)
	callback := func(msg server.Message) bool {
		pong := msg.(*pb.Pong)
		// the peering port must match the destination port
		serviceMap := pong.Services.GetMap()
		if serviceMap == nil {
			return false
		}
		peering := serviceMap[string(service.PeeringKey)]
		if peering == nil || int(peering.GetPort()) != toAddr.Port {
			return false
		}
		// the hash must be correct
		return bytes.Equal(pong.GetReqHash(), hash)
	}

	p.logSend(toAddr, ping)

	return p.Protocol.SendExpectingReply(toAddr, toID, data, pb.MPong, callback)
}

// DiscoveryRequest request known peers from the given target. This method blocks
// until a response is received and the provided peers are returned.
func (p *Protocol) DiscoveryRequest(to *peer.Peer) ([]*peer.Peer, error) {
	if err := p.EnsureVerified(to); err != nil {
		return nil, err
	}

	req := newDiscoveryRequest()
	data := marshal(req)

	// compute the message hash
	hash := server.PacketHash(data)

	peers := make([]*peer.Peer, 0, MaxPeersInResponse)
	callback := func(m server.Message) bool {
		res := m.(*pb.DiscoveryResponse)
		if !bytes.Equal(res.GetReqHash(), hash) {
			return false
		}

		peers = peers[:0]
		for _, protoPeer := range res.GetPeers() {
			if p, _ := peer.FromProto(protoPeer); p != nil {
				peers = append(peers, p)
			}
		}

		return true
	}

	err := backoff.Retry(retryPolicy, func() error {
		p.logSend(to.Address(), req)
		err := <-p.Protocol.SendExpectingReply(to.Address(), to.ID(), data, pb.MDiscoveryResponse, callback)
		if err != nil && !errors.Is(err, server.ErrTimeout) {
			return backoff.Permanent(err)
		}

		return err
	})

	return peers, err
}

// ------ helper functions ------

// hasVerified returns whether the given peer has recently verified the local peer.
func (p *Protocol) hasVerified(id identity.ID, ip net.IP) bool {
	return time.Since(p.loc.Database().LastPing(id, ip)) < PingExpiration
}

func (p *Protocol) logSend(toAddr *net.UDPAddr, msg pb.Message) {
	if logSends {
		p.log.Debugw("send message", "type", msg.Name(), "addr", toAddr)
	}
}

func marshal(msg pb.Message) []byte {
	mType := msg.Type()
	if mType > 0xFF {
		panic("invalid message")
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		panic("invalid message")
	}

	return append([]byte{byte(mType)}, data...)
}

// newPeer creates a new peer that only has a peering service at the given address.
func newPeer(identity *identity.Identity, network string, addr *net.UDPAddr) *peer.Peer {
	services := service.New()
	services.Update(service.PeeringKey, network, addr.Port)

	return peer.NewPeer(identity, addr.IP, services)
}

// ------ Message Constructors ------

func newPing(version uint32, networkID uint32, srcAddr *net.UDPAddr, dstAddr *net.UDPAddr) *pb.Ping {
	return &pb.Ping{
		Version:   version,
		NetworkId: networkID,
		Timestamp: time.Now().Unix(),
		SrcAddr:   srcAddr.IP.String(),
		SrcPort:   uint32(srcAddr.Port),
		DstAddr:   dstAddr.IP.String(),
	}
}

func newPong(dstAddr *net.UDPAddr, reqData []byte, services *service.Record) *pb.Pong {
	return &pb.Pong{
		ReqHash:  server.PacketHash(reqData),
		Services: services.ToProto(),
		DstAddr:  dstAddr.IP.String(),
	}
}

func newDiscoveryRequest() *pb.DiscoveryRequest {
	return &pb.DiscoveryRequest{
		Timestamp: time.Now().Unix(),
	}
}

func newDiscoveryResponse(reqData []byte, list []*peer.Peer) *pb.DiscoveryResponse {
	peers := make([]*peerpb.Peer, 0, len(list))
	for _, p := range list {
		peers = append(peers, p.ToProto())
	}

	return &pb.DiscoveryResponse{
		ReqHash: server.PacketHash(reqData),
		Peers:   peers,
	}
}

// ------ Message Handlers ------

func (p *Protocol) validatePing(fromAddr *net.UDPAddr, m *pb.Ping) bool {
	// check version number
	if m.GetVersion() != p.version {
		p.log.Debugw("invalid message",
			"type", m.Name(),
			"version", m.GetVersion(),
			"want", p.version,
		)

		return false
	}
	// check network identifier
	if m.GetNetworkId() != p.netID {
		p.log.Debugw("invalid message",
			"type", m.Name(),
			"network_id", m.GetNetworkId(),
			"want", p.netID,
		)

		return false
	}
	// check timestamp
	if p.Protocol.IsExpired(m.GetTimestamp()) {
		p.log.Debugw("invalid message",
			"type", m.Name(),
			"timestamp", time.Unix(m.GetTimestamp(), 0),
		)

		return false
	}
	// check that the src_addr is valid
	// currently only the unspecified case is supported, where the source IP address of the packet is used.
	if ip := net.ParseIP(m.GetSrcAddr()); !ip.IsUnspecified() {
		p.log.Debugw("invalid message",
			"type", m.Name(),
			"src_addr", m.GetSrcAddr(),
		)

		return false
	}
	// check that src_port is a valid port number
	if !netutil.IsValidPort(int(m.GetSrcPort())) {
		p.log.Debugw("invalid message",
			"type", m.Name(),
			"src_port", m.GetSrcPort(),
		)

		return false
	}
	// ignore dst_addr and dst_port for now

	p.log.Debugw("valid message",
		"type", m.Name(),
		"addr", fromAddr,
	)

	return true
}

func (p *Protocol) handlePing(s *server.Server, fromAddr *net.UDPAddr, from *identity.Identity, m *pb.Ping, rawData []byte) {
	// create and send the pong response
	pong := newPong(fromAddr, rawData, p.loc.Services().CreateRecord())

	// the destination address uses the source IP address of the packet plus the src_port from the message
	dstAddr := &net.UDPAddr{
		IP:   fromAddr.IP,
		Port: int(m.SrcPort),
	}
	p.logSend(dstAddr, pong)
	s.Send(dstAddr, marshal(pong))

	// if the peer is unknown or expired, send a Ping to verify
	if !p.IsVerified(from.ID(), dstAddr.IP) {
		p.sendPing(dstAddr, from.ID())
	} else if !p.mgr.isKnown(from.ID()) { // add a discovered peer to the manager if it is new but verified
		p.mgr.addDiscoveredPeer(newPeer(from, s.LocalAddr().Network(), dstAddr))
	}

	_ = p.loc.Database().UpdateLastPing(from.ID(), dstAddr.IP, time.Now())
}

func (p *Protocol) validatePong(s *server.Server, fromAddr *net.UDPAddr, fromID identity.ID, m *pb.Pong) bool {
	// there must be a Ping waiting for this pong as a reply
	if !s.IsExpectedReply(fromAddr.IP, fromID, m) {
		p.log.Debugw("invalid message",
			"type", m.Name(),
			"unexpected", fromAddr,
		)

		return false
	}
	// there must a valid number of services
	if numServices := len(m.GetServices().GetMap()); numServices <= 0 || numServices > MaxServices {
		p.log.Debugw("invalid message",
			"type", m.Name(),
			"#peers", numServices,
		)

		return false
	}
	// ignore dst_addr and dst_port for now

	p.log.Debugw("valid message",
		"type", m.Name(),
		"addr", fromAddr,
	)

	return true
}

func (p *Protocol) handlePong(fromAddr *net.UDPAddr, from *identity.Identity, m *pb.Pong) {
	services, err := service.FromProto(m.GetServices())
	if err != nil {
		p.log.Warnf("failed to get services from protocol: %v", err)

		return
	}

	peering := services.Get(service.PeeringKey)
	if peering == nil {
		p.log.Warn("invalid services")

		return
	}

	// create a proper key with these services
	fromPeer := peer.NewPeer(from, fromAddr.IP, services)

	// a valid pong verifies the peer
	_ = p.mgr.addVerifiedPeer(fromPeer)

	// update peer database
	db := p.loc.Database()
	_ = db.UpdateLastPong(from.ID(), fromAddr.IP, time.Now())
	_ = db.UpdatePeer(fromPeer)
}

func (p *Protocol) validateDiscoveryRequest(fromAddr *net.UDPAddr, fromID identity.ID, m *pb.DiscoveryRequest) bool {
	// check Timestamp
	if p.Protocol.IsExpired(m.GetTimestamp()) {
		p.log.Debugw("invalid message",
			"type", m.Name(),
			"timestamp", time.Unix(m.GetTimestamp(), 0),
		)

		return false
	}
	// check whether the sender is verified
	if !p.IsVerified(fromID, fromAddr.IP) {
		p.log.Debugw("invalid message",
			"type", m.Name(),
			"unverified", fromAddr,
		)

		return false
	}

	p.log.Debugw("valid message",
		"type", m.Name(),
		"addr", fromAddr,
	)

	return true
}

func (p *Protocol) handleDiscoveryRequest(s *server.Server, fromID identity.ID, rawData []byte) {
	// get a random list of verified peers
	peers := p.mgr.randomPeers(MaxPeersInResponse, 1)
	res := newDiscoveryResponse(rawData, unwrapPeers(peers))

	from := p.GetVerifiedPeer(fromID)

	p.logSend(from.Address(), res)
	s.Send(from.Address(), marshal(res))
}

func (p *Protocol) validateDiscoveryResponse(s *server.Server, fromAddr *net.UDPAddr, fromID identity.ID, m *pb.DiscoveryResponse) bool {
	// there must be a request waiting for this response
	if !s.IsExpectedReply(fromAddr.IP, fromID, m) {
		p.log.Debugw("invalid message",
			"type", m.Name(),
			"unexpected", fromAddr,
		)

		return false
	}
	// there must not be too many peers
	if len(m.GetPeers()) > MaxPeersInResponse {
		p.log.Debugw("invalid message",
			"type", m.Name(),
			"#peers", len(m.GetPeers()),
		)

		return false
	}

	p.log.Debugw("valid message",
		"type", m.Name(),
		"addr", fromAddr,
	)

	return true
}
