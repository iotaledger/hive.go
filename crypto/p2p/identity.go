package p2p

import (
	"encoding/hex"
	"os"

	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"

	"github.com/iotaledger/hive.go/crypto/pem"
	"github.com/iotaledger/hive.go/ierrors"
)

// LoadOrCreateIdentityPrivateKey loads an existing Ed25519 based identity private key
// or creates a new one and stores it as a PEM file in the p2p store folder.
func LoadOrCreateIdentityPrivateKey(privKeyFilePath string, identityPrivKey string) (libp2pcrypto.PrivKey, bool, error) {
	privKeyFromConfig, err := ParseLibp2pEd25519PrivateKeyFromString(identityPrivKey)
	if err != nil {
		if ierrors.Is(err, ErrPrivKeyInvalid) {
			return nil, false, ierrors.New("configuration contains an invalid private key")
		}

		if !ierrors.Is(err, ErrNoPrivKeyFound) {
			return nil, false, ierrors.Wrap(err, "unable to parse private key from config")
		}
	}

	_, err = os.Stat(privKeyFilePath)
	switch {
	case err == nil || os.IsExist(err):
		// private key already exists, load and return it
		privKey, err := pem.ReadEd25519PrivateKeyFromPEMFile(privKeyFilePath)
		if err != nil {
			return nil, false, ierrors.Wrap(err, "unable to load Ed25519 private key for peer identity")
		}

		libp2pPrivKey, err := Ed25519PrivateKeyToLibp2pPrivateKey(privKey)
		if err != nil {
			return nil, false, err
		}

		if privKeyFromConfig != nil && !privKeyFromConfig.Equals(libp2pPrivKey) {
			storedPrivKeyBytes, err := libp2pcrypto.MarshalPrivateKey(libp2pPrivKey)
			if err != nil {
				return nil, false, ierrors.Wrap(err, "unable to marshal stored Ed25519 private key for peer identity")
			}
			configPrivKeyBytes, err := libp2pcrypto.MarshalPrivateKey(privKeyFromConfig)
			if err != nil {
				return nil, false, ierrors.Wrap(err, "unable to marshal configured Ed25519 private key for peer identity")
			}

			return nil, false, ierrors.Errorf("stored Ed25519 private key (%s) for peer identity doesn't match private key in config (%s)", hex.EncodeToString(storedPrivKeyBytes), hex.EncodeToString(configPrivKeyBytes))
		}

		return libp2pPrivKey, false, nil

	case os.IsNotExist(err):
		var libp2pPrivKey libp2pcrypto.PrivKey

		if privKeyFromConfig != nil {
			libp2pPrivKey = privKeyFromConfig
		} else {
			// private key does not exist, create a new one
			libp2pPrivKey, _, err = libp2pcrypto.GenerateKeyPair(libp2pcrypto.Ed25519, -1)
			if err != nil {
				return nil, false, ierrors.Wrap(err, "unable to generate Ed25519 private key for peer identity")
			}
		}

		privKey, err := Libp2pPrivateKeyToEd25519PrivateKey(libp2pPrivKey)
		if err != nil {
			return nil, false, err
		}

		if err := pem.WriteEd25519PrivateKeyToPEMFile(privKeyFilePath, privKey); err != nil {
			return nil, false, ierrors.Wrap(err, "unable to store private key file for peer identity")
		}

		return libp2pPrivKey, true, nil

	default:
		return nil, false, ierrors.Wrapf(err, "unable to check private key file for peer identity (%s)", privKeyFilePath)
	}
}
