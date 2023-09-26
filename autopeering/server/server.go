package server

import (
	"container/list"
	"net"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/izuc/zipp.foundation/autopeering/netutil"
	"github.com/izuc/zipp.foundation/autopeering/peer"
	pb "github.com/izuc/zipp.foundation/autopeering/server/proto"
	"github.com/izuc/zipp.foundation/crypto/identity"
	"github.com/izuc/zipp.foundation/lo"
	"github.com/izuc/zipp.foundation/logger"
	"github.com/izuc/zipp.foundation/runtime/timeutil"
)

const (
	// ResponseTimeout specifies the time limit after which a response must have been received.
	ResponseTimeout = 500 * time.Millisecond
)

// NetConn defines the interface required for a connection.
type NetConn interface {
	// Close closes the connection.
	Close() error

	// LocalAddr returns the local network address.
	LocalAddr() net.Addr

	// ReadFromUDP acts like ReadFrom but returns a UDPAddr.
	ReadFromUDP([]byte) (int, *net.UDPAddr, error)
	// WriteToUDP acts like WriteTo but takes a UDPAddr.
	WriteToUDP([]byte, *net.UDPAddr) (int, error)
}

// Server offers the functionality to start a server that handles requests and responses from peers.
type Server struct {
	local    *peer.Local
	conn     NetConn
	handlers []Handler
	log      *logger.Logger
	network  string

	closeOnce sync.Once
	wg        sync.WaitGroup

	addReplyMatcher chan *replyMatcher
	replyReceived   chan reply
	closing         chan struct{} // if this channel gets closed all pending waits should terminate
}

// a replyMatcher stores the information required to identify and react to an expected replay.
type replyMatcher struct {
	// fromIP must match the sender of the reply
	fromIP net.IP
	// fromID must match the sender ID
	fromID identity.ID
	// mtype must match the type of the reply
	mtype MType

	// deadline when the request must complete
	deadline time.Time

	// callback is called when a matching reply arrives
	// If it returns true, the reply is acceptable.
	callback func(msg Message) bool

	// errc receives nil when the callback indicates completion or an
	// error if no further reply is received within the timeout
	errc chan error
}

// reply is a reply packet from a certain peer.
type reply struct {
	fromIP         net.IP
	fromID         identity.ID
	msg            Message     // the actual reply message
	matchedRequest chan<- bool // a matching request is indicated via this channel
}

// Serve starts a new peer server using the given transport layer for communication.
// Sent data is signed using the identity of the local peer,
// received data with a valid peer signature is handled according to the provided Handler.
func Serve(local *peer.Local, conn NetConn, log *logger.Logger, h ...Handler) *Server {
	srv := &Server{
		local:           local,
		conn:            conn,
		handlers:        h,
		log:             log,
		network:         local.Network(),
		addReplyMatcher: make(chan *replyMatcher),
		replyReceived:   make(chan reply),
		closing:         make(chan struct{}),
	}

	srv.wg.Add(2)
	go srv.replyLoop()
	go srv.readLoop()

	log.Debugw("server started",
		"network", srv.LocalAddr().Network(),
		"address", srv.LocalAddr().String(),
		"#handlers", len(h))

	return srv
}

// Close shuts down the server.
func (s *Server) Close() {
	s.closeOnce.Do(func() {
		close(s.closing)
		s.conn.Close()
		s.wg.Wait()
	})
}

// Local returns the the local peer.
func (s *Server) Local() *peer.Local {
	return s.local
}

// LocalAddr returns the address of the local peer in string form.
func (s *Server) LocalAddr() *net.UDPAddr {
	return s.conn.LocalAddr().(*net.UDPAddr)
}

// Send sends a message to the given address.
func (s *Server) Send(toAddr *net.UDPAddr, data []byte) {
	pkt := s.encode(data)
	s.write(pkt, toAddr)
}

// SendExpectingReply sends a message to the given address and tells the Server
// to expect a reply message with the given specifications.
// If eventually nil is returned, a matching message was received.
func (s *Server) SendExpectingReply(toAddr *net.UDPAddr, toID identity.ID, data []byte, replyType MType, callback func(Message) bool) <-chan error {
	errc := s.expectReply(toAddr.IP, toID, replyType, callback)
	s.Send(toAddr, data)

	return errc
}

// expectReply tells the Server to expect a reply message with the given specifications.
// If eventually nil is returned, a matching message was received.
func (s *Server) expectReply(fromIP net.IP, fromID identity.ID, mtype MType, callback func(Message) bool) <-chan error {
	ch := make(chan error, 1)
	m := &replyMatcher{fromIP: fromIP, fromID: fromID, mtype: mtype, callback: callback, errc: ch}
	select {
	case s.addReplyMatcher <- m:
	case <-s.closing:
		ch <- ErrClosed
	}

	return ch
}

// IsExpectedReply checks whether the given testMessage matches an expected reply added with SendExpectingReply.
func (s *Server) IsExpectedReply(fromIP net.IP, fromID identity.ID, msg Message) bool {
	matched := make(chan bool, 1)
	select {
	case s.replyReceived <- reply{fromIP, fromID, msg, matched}:
		// wait for matcher and return whether a matching request was found
		return <-matched
	case <-s.closing:
		return false
	}
}

// Loop checking for matching replies.
func (s *Server) replyLoop() {
	defer s.wg.Done()

	var (
		matcherList = list.New()
		timeout     = time.NewTimer(0)
	)
	defer timeutil.CleanupTimer(timeout)

	<-timeout.C // ignore first timeout

	for {

		// Set the timer so that it fires when the next reply expires
		if e := matcherList.Front(); e != nil {
			// the first element always has the closest deadline
			m := e.Value.(*replyMatcher)
			timeout.Reset(time.Until(m.deadline))
		} else {
			timeout.Stop()
		}

		select {

		// add a new matcher to the list
		case s := <-s.addReplyMatcher:
			s.deadline = time.Now().Add(ResponseTimeout)
			matcherList.PushBack(s)

		// on reply received, check all matchers for fits
		case r := <-s.replyReceived:
			matched := false
			for e := matcherList.Front(); e != nil; e = e.Next() {
				m := e.Value.(*replyMatcher)
				if m.mtype == r.msg.Type() && m.fromID == r.fromID && m.fromIP.Equal(r.fromIP) {
					if m.callback(r.msg) {
						// request has been matched
						matched = true
						m.errc <- nil
						matcherList.Remove(e)
					}
				}
			}
			r.matchedRequest <- matched

		// on timeout, check for expired matchers
		case <-timeout.C:
			now := time.Now()

			// notify and remove any expired matchers
			for e := matcherList.Front(); e != nil; e = e.Next() {
				m := e.Value.(*replyMatcher)
				if now.After(m.deadline) || now.Equal(m.deadline) {
					m.errc <- ErrTimeout
					matcherList.Remove(e)
				}
			}

		// on close, notice all the matchers
		case <-s.closing:
			for e := matcherList.Front(); e != nil; e = e.Next() {
				e.Value.(*replyMatcher).errc <- ErrClosed
			}

			return

		}
	}
}

func (s *Server) write(pkt *pb.Packet, toAddr *net.UDPAddr) {
	b, err := proto.Marshal(pkt)
	if err != nil {
		s.log.Error("marshal error", "err", err)

		return
	}
	if l := len(b); l > MaxPacketSize {
		s.log.Error("packet too large", "size", l, "max", MaxPacketSize)

		return
	}

	_, err = s.conn.WriteToUDP(b, toAddr)
	if err != nil {
		s.log.Debugw("failed to write packet", "addr", toAddr, "err", err)
	}
}

// encodes a message as a data packet that can be written.
func (s *Server) encode(data []byte) *pb.Packet {
	if len(data) == 0 {
		panic("server: no data")
	}

	return &pb.Packet{
		PublicKey: lo.PanicOnErr(s.local.PublicKey().Bytes()),
		Signature: lo.PanicOnErr(s.local.Sign(data).Bytes()),
		Data:      append([]byte(nil), data...),
	}
}

func (s *Server) readLoop() {
	defer s.wg.Done()

	buffer := make([]byte, MaxPacketSize)
	for {
		n, fromAddr, err := s.conn.ReadFromUDP(buffer)
		if netutil.IsTemporaryError(err) {
			// ignore temporary read errors.
			s.log.Debugw("temporary read error", "err", err)

			continue
		}
		// return from the loop on all other errors
		if err != nil {
			// The error that is returned for an operation on a closed network connection is not exported.
			// This is the only way to check for the error. See issues #4373 and #19252.
			if !strings.Contains(err.Error(), "use of closed network connection") {
				s.log.Warnw("read error", "err", err)
			}
			s.log.Debug("reading stopped")

			return
		}

		pkt := new(pb.Packet)
		if err := proto.Unmarshal(buffer[:n], pkt); err != nil {
			s.log.Debugw("bad packet", "from", fromAddr, "err", err)

			continue
		}
		if err := s.handlePacket(pkt, fromAddr); err != nil {
			s.log.Debugw("failed to handle packet", "from", fromAddr, "err", err)
		}
	}
}

func (s *Server) handlePacket(pkt *pb.Packet, fromAddr *net.UDPAddr) error {
	data, from, err := decode(pkt)
	if err != nil {
		return err
	}

	for _, handler := range s.handlers {
		ok, err := handler.HandleMessage(s, fromAddr, from, data)
		if ok {
			return err
		}
	}

	return ErrInvalidMessage
}

func decode(pkt *pb.Packet) ([]byte, *identity.Identity, error) {
	if len(pkt.GetData()) == 0 {
		return nil, nil, ErrNoMessage
	}

	key, err := peer.RecoverKeyFromSignedData(pkt)
	if err != nil {
		return nil, nil, err
	}

	return pkt.GetData(), identity.New(key), nil
}
