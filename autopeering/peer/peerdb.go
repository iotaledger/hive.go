package peer

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/hive.go/kvstore"
)

const (
	// remove peers from DB, when the last received ping was older than this
	peerExpiration = 24 * time.Hour
	// interval in which expired peers are checked
	cleanupInterval = time.Hour

	// number of peers used for bootstrapping
	seedCount = 10
	// time after which potential seed peers should expire
	seedExpiration = 5 * 24 * time.Hour
)

// DB is the peer database, storing previously seen peers and any collected properties of them.
type DB struct {
	store kvstore.KVStore
}

// Keys in the node database.
const (
	dbNodePrefix  = "n:"     // Identifier to prefix node entries with
	dbLocalPrefix = "local:" // Identifier to prefix local entries

	// These fields are stored per ID and address. Use nodeFieldKey to create those keys.
	dbNodePing = "lastping"
	dbNodePong = "lastpong"

	// Local information is keyed by ID only. Use localFieldKey to create those keys.
	dbLocalKey = "key"
)

// NewDB creates a new peer database.
func NewDB(store kvstore.KVStore) (*DB, error) {
	pDB := &DB{
		store: store,
	}
	err := pDB.init()
	if err != nil {
		return nil, err
	}
	return pDB, nil
}

func (db *DB) init() error {
	// get all peers in the DB
	peers := db.getPeers(0)

	for _, p := range peers {
		// if they dont have an associated pong, give them a grace period
		if db.LastPong(p.ID(), p.IP()).Unix() == 0 {
			if err := db.setPeerWithTTL(p, cleanupInterval); err != nil {
				return err
			}
		}
	}
	return nil
}

// nodeKey returns the database key for a node record.
func nodeKey(id identity.ID) []byte {
	return append([]byte(dbNodePrefix), id.Bytes()...)
}

// nodeFieldKey returns the database key for a node metadata field.
func nodeFieldKey(id identity.ID, ip net.IP, field string) []byte {
	return bytes.Join([][]byte{nodeKey(id), ip.To16(), []byte(field)}, []byte{':'})
}

func localFieldKey(field string) []byte {
	return append([]byte(dbLocalPrefix), []byte(field)...)
}

func parseInt64(blob []byte) int64 {
	val, read := binary.Varint(blob)
	if read <= 0 {
		return 0
	}
	return val
}

// getInt64 retrieves an integer associated with a particular key.
func (db *DB) getInt64(key []byte) int64 {
	value, err := db.store.Get(key)
	if err != nil {
		return 0
	}
	return parseInt64(value)
}

// setInt64 stores an integer in the given key.
func (db *DB) setInt64(key []byte, n int64) error {
	blob := make([]byte, binary.MaxVarintLen64)
	blob = blob[:binary.PutVarint(blob, n)]
	return db.store.Set(key, blob) //FIXME: handle TTL with peerExpiration
}

// LocalPrivateKey returns the private key stored in the database or creates a new one.
func (db *DB) LocalPrivateKey() (privateKey ed25519.PrivateKey, err error) {
	value, err := db.store.Get(localFieldKey(dbLocalKey))
	if err == kvstore.ErrKeyNotFound {
		key, genErr := ed25519.GeneratePrivateKey()
		if genErr == nil {
			err = db.UpdateLocalPrivateKey(key)
		}
		return key, err
	}
	if err != nil {
		return
	}

	copy(privateKey[:], value)
	return
}

// UpdateLocalPrivateKey stores the provided key in the database.
func (db *DB) UpdateLocalPrivateKey(key ed25519.PrivateKey) error {
	return db.store.Set(localFieldKey(dbLocalKey), key.Bytes())
}

// LastPing returns that property for the given peer ID and address.
func (db *DB) LastPing(id identity.ID, ip net.IP) time.Time {
	return time.Unix(db.getInt64(nodeFieldKey(id, ip, dbNodePing)), 0)
}

// UpdateLastPing updates that property for the given peer ID and address.
func (db *DB) UpdateLastPing(id identity.ID, ip net.IP, t time.Time) error {
	return db.setInt64(nodeFieldKey(id, ip, dbNodePing), t.Unix())
}

// LastPong returns that property for the given peer ID and address.
func (db *DB) LastPong(id identity.ID, ip net.IP) time.Time {
	return time.Unix(db.getInt64(nodeFieldKey(id, ip, dbNodePong)), 0)
}

// UpdateLastPong updates that property for the given peer ID and address.
func (db *DB) UpdateLastPong(id identity.ID, ip net.IP, t time.Time) error {
	return db.setInt64(nodeFieldKey(id, ip, dbNodePong), t.Unix())
}

func (db *DB) setPeerWithTTL(p *Peer, ttl time.Duration) error {
	data, err := p.Marshal()
	if err != nil {
		return err
	}
	return db.store.Set(nodeKey(p.ID()), data) //FIXME: handle TTL with ttl
}

// UpdatePeer updates a peer in the database.
func (db *DB) UpdatePeer(p *Peer) error {
	return db.setPeerWithTTL(p, peerExpiration)
}

func parsePeer(data []byte) (*Peer, error) {
	p, err := Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Peer retrieves a peer from the database.
func (db *DB) Peer(id identity.ID) (*Peer, error) {
	data, err := db.store.Get(nodeKey(id))
	if err != nil {
		return nil, err
	}
	return parsePeer(data)
}

func randomSubset(peers []*Peer, m int) []*Peer {
	if len(peers) <= m {
		return peers
	}

	result := make([]*Peer, 0, m)
	for i, p := range peers {
		if rand.Intn(len(peers)-i) < m-len(result) {
			result = append(result, p)
		}
	}
	return result
}

func (db *DB) getPeers(maxAge time.Duration) (peers []*Peer) {
	now := time.Now()

	err := db.store.Iterate([]kvstore.KeyPrefix{[]byte(dbNodePrefix)}, func(key kvstore.Key, value kvstore.Value) bool {
		keyWithoutPrefix := key[len(dbNodePrefix):]
		var id identity.ID
		if len(keyWithoutPrefix) != len(id) {
			return true
		}
		copy(id[:], keyWithoutPrefix)

		p, err := parsePeer(value)
		if err != nil || p.ID() != id {
			return true
		}
		if maxAge > 0 && now.Sub(db.LastPong(p.ID(), p.IP())) > maxAge {
			return true
		}

		peers = append(peers, p)
		return true
	})

	if err != nil {
		return nil
	}
	return peers
}

// SeedPeers retrieves random nodes to be used as potential bootstrap peers.
func (db *DB) SeedPeers() []*Peer {
	// get all stored peers and select subset
	peers := db.getPeers(0)

	return randomSubset(peers, seedCount)
}

// PersistSeeds assures that potential bootstrap peers are not garbage collected.
// It returns the number of peers that have been persisted.
func (db *DB) PersistSeeds() int {
	// randomly select potential bootstrap peers
	peers := randomSubset(db.getPeers(peerExpiration), seedCount)

	for i, p := range peers {
		err := db.setPeerWithTTL(p, seedExpiration)
		if err != nil {
			return i
		}
	}
	return len(peers)
}
