package pem

import (
	"bytes"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/runtime/ioutils"
)

// ReadEd25519PrivateKeyFromPEMFile reads an Ed25519 private key from a file with PEM format.
func ReadEd25519PrivateKeyFromPEMFile(filepath string) (ed25519.PrivateKey, error) {
	pemPrivateBlockBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, ierrors.Wrap(err, "unable to read private key")
	}

	pemPrivateBlock, _ := pem.Decode(pemPrivateBlockBytes)
	if pemPrivateBlock == nil {
		return nil, ierrors.New("unable to decode private key")
	}

	cryptoPrivKey, err := x509.ParsePKCS8PrivateKey(pemPrivateBlock.Bytes)
	if err != nil {
		return nil, ierrors.Wrap(err, "unable to parse private key")
	}

	privKey, ok := cryptoPrivKey.(ed25519.PrivateKey)
	if !ok {
		return nil, ierrors.New("unable to type assert private key")
	}

	return privKey, nil
}

// WriteEd25519PrivateKeyToPEMFile stores an Ed25519 private key to a file with PEM format.
func WriteEd25519PrivateKeyToPEMFile(filepath string, privateKey ed25519.PrivateKey) error {
	if err := ioutils.CreateDirectory(path.Dir(filepath), 0o700); err != nil {
		return ierrors.Wrap(err, "unable to store private key")
	}

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return ierrors.Wrap(err, "unable to marshal private key")
	}

	pemPrivateBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	var pemBuffer bytes.Buffer
	if err := pem.Encode(&pemBuffer, pemPrivateBlock); err != nil {
		return ierrors.Wrap(err, "unable to encode private key")
	}

	if err := ioutils.WriteToFile(filepath, pemBuffer.Bytes(), 0660); err != nil {
		return ierrors.Wrap(err, "unable to write private key")
	}

	return nil
}
