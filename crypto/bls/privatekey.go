package bls

import (
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/mr-tron/base58"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"golang.org/x/xerrors"
)

// PrivateKey is the type of BLS private keys.
type PrivateKey struct {
	Scalar kyber.Scalar
}

// PrivateKeyFromBytes creates a PrivateKey from the given bytes.
func PrivateKeyFromBytes(bytes []byte) (privateKey PrivateKey, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if privateKey, err = PrivateKeyFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse PrivateKey from MarshalUtil: %w", err)
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// PrivateKeyFromBase58EncodedString creates a PrivateKey from a base58 encoded string.
func PrivateKeyFromBase58EncodedString(base58String string) (privateKey PrivateKey, err error) {
	bytes, err := base58.Decode(base58String)
	if err != nil {
		err = xerrors.Errorf("error while decoding base58 encoded PrivateKey (%v): %w", err, ErrBase58DecodeFailed)
		return
	}

	if privateKey, _, err = PrivateKeyFromBytes(bytes); err != nil {
		err = xerrors.Errorf("failed to parse PrivateKey from bytes: %w", err)
		return
	}

	return
}

// PrivateKeyFromMarshalUtil unmarshals a PrivateKey using a MarshalUtil (for easier unmarshaling).
func PrivateKeyFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (privateKey PrivateKey, err error) {
	bytes, err := marshalUtil.ReadBytes(PrivateKeySize)
	if err != nil {
		err = xerrors.Errorf("failed to read PrivateKey bytes (%v): %w", err, ErrParseBytesFailed)
		return
	}

	if err = privateKey.Scalar.UnmarshalBinary(bytes); err != nil {
		err = xerrors.Errorf("failed to unmarshal PrivateKey (%v): %w", err, ErrParseBytesFailed)
		return
	}

	return
}

// PrivateKeyFromRandomness generates a new random PrivateKey.
func PrivateKeyFromRandomness() (privateKey PrivateKey) {
	privateKey.Scalar, _ = bdn.NewKeyPair(blsSuite, randomness)

	return
}

// PublicKey returns the PublicKey corresponding to the PrivateKey.
func (p PrivateKey) PublicKey() PublicKey {
	return PublicKey{
		Point: blsSuite.G2().Point().Mul(p.Scalar, nil),
	}
}

// Sign signs the message and returns a SignatureWithPublicKey.
func (p PrivateKey) Sign(data []byte) (signatureWithPublicKey SignatureWithPublicKey, err error) {
	sig, err := bdn.Sign(blsSuite, p.Scalar, data)
	if err != nil {
		err = xerrors.Errorf("failed to sign data (%v): %w", err, ErrBLSFailed)
		return
	}

	signatureWithPublicKey.PublicKey = p.PublicKey()
	copy(signatureWithPublicKey.Signature[:], sig)

	return
}

// Bytes returns a marshaled version of the PrivateKey.
func (p PrivateKey) Bytes() (bytes []byte) {
	bytes, err := p.Scalar.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return
}

// Base58 returns a base58 encoded version of the PrivateKey.
func (p PrivateKey) Base58() string {
	return base58.Encode(p.Bytes())
}

// String returns a human readable version of the PrivateKey (base58 encoded).
func (p PrivateKey) String() string {
	return p.Base58()
}
