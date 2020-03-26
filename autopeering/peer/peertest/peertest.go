// Package peertest provides utilities for writing tests with the peer package.
package peertest

import (
	"log"
	"math/rand"
	"net"
	"strconv"

	"github.com/iotaledger/hive.go/autopeering/peer"
	"github.com/iotaledger/hive.go/autopeering/peer/service"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/identity"
)

// NewPeer creates a new peer for tests.
func NewPeer(network string, ip string, port int) *peer.Peer {
	services := service.New()
	services.Update(service.PeeringKey, network, port)
	key := ed25519.PublicKey{}
	copy(key[:], net.JoinHostPort(ip, strconv.Itoa(port)))
	return peer.NewPeer(identity.NewIdentity(key), net.ParseIP(ip), services)
}

// NewLocal crates a new local for tests.
func NewLocal(network string, ip net.IP, port int, db *peer.DB) *peer.Local {
	services := service.New()
	services.Update(service.PeeringKey, network, port)
	local, err := peer.NewLocal(ip, services, db, randomSeed())
	if err != nil {
		log.Panic(err)
	}
	return local
}

func randomSeed() []byte {
	seed := make([]byte, ed25519.SeedSize)
	rand.Read(seed)
	return seed
}
