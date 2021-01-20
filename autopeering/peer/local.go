package peer

import (
	"net"
	"sync"

	"github.com/iotaledger/hive.go/autopeering/arrow"
	"github.com/iotaledger/hive.go/autopeering/peer/service"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/identity"
)

// Local defines the struct of a local peer
type Local struct {
	*Peer
	identity *identity.LocalIdentity
	db       *DB

	// everything below is protected by a lock
	mu            sync.RWMutex
	serviceRecord *service.Record
	arrow         *arrow.ArRow
}

// newLocal creates a new local peer.
func newLocal(key ed25519.PrivateKey, ip net.IP, serviceRecord *service.Record, db *DB) *Local {
	id := identity.New(key.Public())

	return &Local{
		Peer:          NewPeer(id, ip, serviceRecord),
		identity:      identity.NewLocalIdentityWithIdentity(id, key),
		db:            db,
		serviceRecord: serviceRecord,
	}
}

// NewLocal creates a new local peer linked to the provided store.
// If an optional seed is provided, the seed is used to generate the private key. Without a seed,
// the provided key is loaded from the provided database and generated if not stored there.
func NewLocal(ip net.IP, serviceRecord *service.Record, db *DB, seed ...[]byte) (*Local, error) {
	var key ed25519.PrivateKey
	if len(seed) > 0 {
		key = ed25519.PrivateKeyFromSeed(seed[0])
		if db != nil {
			if err := db.UpdateLocalPrivateKey(key); err != nil {
				return nil, err
			}
		}
	} else {
		var err error
		key, err = db.LocalPrivateKey()
		if err != nil {
			return nil, err
		}
	}

	return newLocal(key, ip, serviceRecord, db), nil
}

// Database returns the node database associated with the local peer.
func (l *Local) Database() *DB {
	return l.db
}

// UpdateService updates the endpoint address of the given local service.
func (l *Local) UpdateService(key service.Key, network string, port int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// update the service in the read protected map
	l.serviceRecord.Update(key, network, port)

	// create a new peer with the corresponding services
	l.Peer = NewPeer(l.identity.Identity, l.IP(), l.serviceRecord)

	return nil
}

// GetArRow returns current Ar values
func (l *Local) GetArRow() *arrow.ArRow {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.arrow
}

// SetArRow sets current Ar values
func (l *Local) SetArRow(arrow *arrow.ArRow) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.arrow = arrow
}

// Sign signs a message using the node's LocalIdentity.
func (l *Local) Sign(message []byte) ed25519.Signature {
	return l.identity.Sign(message)
}

// LocalIdentity returns the local identity
func (l *Local) LocalIdentity() *identity.LocalIdentity {
	return l.identity
}
