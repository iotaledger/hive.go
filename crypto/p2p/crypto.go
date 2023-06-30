package p2p

import (
	"crypto/ed25519"

	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"

	"github.com/iotaledger/hive.go/crypto"
	"github.com/iotaledger/hive.go/ierrors"
)

var (
	ErrPrivKeyInvalid = ierrors.New("invalid private key")
	ErrNoPrivKeyFound = ierrors.New("no private key found")
)

// ParseLibp2pEd25519PrivateKeyFromString parses an Ed25519 private key from a hex encoded string.
func ParseLibp2pEd25519PrivateKeyFromString(identityPrivKey string) (libp2pcrypto.PrivKey, error) {
	if identityPrivKey == "" {
		return nil, ErrNoPrivKeyFound
	}

	privKey, err := crypto.ParseEd25519PrivateKeyFromString(identityPrivKey)
	if err != nil {
		return nil, ierrors.Wrap(ErrPrivKeyInvalid, "unable to parse private key")
	}

	libp2pPrivKey, err := Ed25519PrivateKeyToLibp2pPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	return libp2pPrivKey, nil
}

func Ed25519PrivateKeyToLibp2pPrivateKey(privKey ed25519.PrivateKey) (libp2pcrypto.PrivKey, error) {
	libp2pPrivKey, _, err := libp2pcrypto.KeyPairFromStdKey(&privKey)
	if err != nil {
		return nil, ierrors.Wrap(err, "unable to unmarshal private key")
	}

	return libp2pPrivKey, nil
}

func Libp2pPrivateKeyToEd25519PrivateKey(libp2pPrivKey libp2pcrypto.PrivKey) (ed25519.PrivateKey, error) {
	cryptoPrivKey, err := libp2pcrypto.PrivKeyToStdKey(libp2pPrivKey)
	if err != nil {
		return nil, ierrors.Wrap(err, "unable to convert private key")
	}

	privKey, ok := cryptoPrivKey.(*ed25519.PrivateKey)
	if !ok {
		return nil, ierrors.New("unable to type assert private key")
	}

	return *privKey, nil
}
