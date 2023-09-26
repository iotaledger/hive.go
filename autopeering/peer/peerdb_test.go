package peer

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/izuc/zipp.foundation/crypto/ed25519"
	"github.com/izuc/zipp.foundation/kvstore/mapdb"
)

func TestPeerDB(t *testing.T) {
	db, err := NewDB(mapdb.NewMapDB())
	require.NoError(t, err)
	defer db.Close()
	p := newTestPeer("test")

	t.Run("LocalPrivateKey", func(t *testing.T) {
		key, err := db.LocalPrivateKey()
		require.NoError(t, err)

		assert.EqualValues(t, ed25519.PrivateKeySize, len(key))
		assert.EqualValues(t, ed25519.PublicKeySize, len(key.Public()))

		err = db.UpdateLocalPrivateKey(key)
		require.NoError(t, err)

		key2, err := db.LocalPrivateKey()
		require.NoError(t, err)

		assert.Equal(t, key, key2)
	})

	t.Run("Peer", func(t *testing.T) {
		err := db.UpdatePeer(p)
		require.NoError(t, err)

		dbPeer, err := db.Peer(p.ID())
		assert.NoError(t, err)
		assert.Equal(t, p, dbPeer)
	})

	t.Run("LastPing", func(t *testing.T) {
		now := time.Now()
		err := db.UpdateLastPing(p.ID(), p.IP(), now)
		require.NoError(t, err)

		assert.Equal(t, now.Unix(), db.LastPing(p.ID(), p.IP()).Unix())
	})

	t.Run("LastPong", func(t *testing.T) {
		now := time.Now()
		err := db.UpdateLastPong(p.ID(), p.IP(), now)
		require.NoError(t, err)

		assert.Equal(t, now.Unix(), db.LastPong(p.ID(), p.IP()).Unix())
	})

	t.Run("getPeers", func(t *testing.T) {
		time.Sleep(time.Second) // let the old peers expire

		newPeer := newTestPeer("new")
		assert.NoError(t, db.UpdatePeer(newPeer))
		assert.NoError(t, db.UpdateLastPong(newPeer.ID(), newPeer.IP(), time.Now()))

		peers := db.getPeers(time.Second)
		assert.ElementsMatch(t, []*Peer{newPeer}, peers)
	})

	t.Run("SeedPeers", func(t *testing.T) {
		for i := 0; i < seedCount+1; i++ {
			p := newTestPeer(fmt.Sprintf("SeedPeers%0d", i))
			assert.NoError(t, db.UpdatePeer(p))
			assert.NoError(t, db.UpdateLastPong(p.ID(), p.IP(), time.Now()))
		}

		peers := db.SeedPeers()
		assert.EqualValues(t, seedCount, len(peers))
	})

	t.Run("expireNodes", func(t *testing.T) {
		newPeer := newTestPeer("new")
		assert.NoError(t, db.UpdatePeer(newPeer))
		assert.NoError(t, db.UpdateLastPong(newPeer.ID(), newPeer.IP(), time.Now()))
		assert.Contains(t, db.getPeers(0), newPeer)

		expPeer := newTestPeer("expired")
		assert.NoError(t, db.UpdatePeer(expPeer))
		assert.NoError(t, db.UpdateLastPing(expPeer.ID(), expPeer.IP(), time.Now()))
		assert.NoError(t, db.UpdateLastPong(expPeer.ID(), expPeer.IP(), time.Now().Add(-peerExpiration).Add(-time.Second)))
		assert.Contains(t, db.getPeers(0), expPeer)

		require.NoError(t, db.expireNodes())
		assert.Contains(t, db.getPeers(0), newPeer)
		assert.NotContains(t, db.getPeers(0), expPeer)
	})
}
