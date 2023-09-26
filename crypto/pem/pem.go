package pem

import (
	"bytes"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"

	"github.com/izuc/zipp.foundation/runtime/ioutils"
)

// ReadEd25519PrivateKeyFromPEMFile reads an Ed25519 private key from a file with PEM format.
func ReadEd25519PrivateKeyFromPEMFile(filepath string) (ed25519.PrivateKey, error) {

	pemPrivateBlockBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	pemPrivateBlock, _ := pem.Decode(pemPrivateBlockBytes)
	if pemPrivateBlock == nil {
		return nil, errors.New("unable to decode private key")
	}

	cryptoPrivKey, err := x509.ParsePKCS8PrivateKey(pemPrivateBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	privKey, ok := cryptoPrivKey.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("unable to type assert private key")
	}

	return privKey, nil
}

// WriteEd25519PrivateKeyToPEMFile stores an Ed25519 private key to a file with PEM format.
func WriteEd25519PrivateKeyToPEMFile(filepath string, privateKey ed25519.PrivateKey) error {

	if err := ioutils.CreateDirectory(path.Dir(filepath), 0o700); err != nil {
		return fmt.Errorf("unable to store private key: %w", err)
	}

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("unable to marshal private key: %w", err)
	}

	pemPrivateBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	var pemBuffer bytes.Buffer
	if err := pem.Encode(&pemBuffer, pemPrivateBlock); err != nil {
		return fmt.Errorf("unable to encode private key: %w", err)
	}

	if err := ioutils.WriteToFile(filepath, pemBuffer.Bytes(), 0660); err != nil {
		return fmt.Errorf("unable to write private key: %w", err)
	}

	return nil
}
