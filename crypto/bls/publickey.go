package bls

import (
	"github.com/mr-tron/base58"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/sign/bdn"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

// PublicKey is the type of BLS public keys.
type PublicKey struct {
	Point kyber.Point
}

// PublicKeyFromBytes creates a PublicKey from the given bytes.
func PublicKeyFromBytes(bytes []byte) (publicKey PublicKey, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if publicKey, err = PublicKeyFromMarshalUtil(marshalUtil); err != nil {
		err = ierrors.Wrap(err, "failed to parse PublicKey from MarshalUtil")
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// PublicKeyFromBase58EncodedString creates a PublicKey from a base58 encoded string.
func PublicKeyFromBase58EncodedString(base58String string) (publicKey PublicKey, err error) {
	bytes, err := base58.Decode(base58String)
	if err != nil {
		err = ierrors.Wrapf(ErrBase58DecodeFailed, "error while decoding base58 encoded PublicKey: %w", err)

		return
	}

	if publicKey, _, err = PublicKeyFromBytes(bytes); err != nil {
		err = ierrors.Wrap(err, "failed to parse PublicKey from bytes")

		return
	}

	return
}

// PublicKeyFromMarshalUtil unmarshals a PublicKey using a MarshalUtil (for easier unmarshalling).
func PublicKeyFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (publicKey PublicKey, err error) {
	bytes, err := marshalUtil.ReadBytes(PublicKeySize)
	if err != nil {
		err = ierrors.Wrapf(ErrParseBytesFailed, "failed to read PublicKey bytes: %w", err)

		return
	}
	publicKey.Point = blsSuite.G2().Point()
	if err = publicKey.Point.UnmarshalBinary(bytes); err != nil {
		err = ierrors.Wrapf(ErrParseBytesFailed, "failed to unmarshal PublicKey: %w", err)

		return
	}

	return
}

// SignatureValid reports whether the signature is valid for the given data.
func (p PublicKey) SignatureValid(data []byte, signature Signature) bool {
	return bdn.Verify(blsSuite, p.Point, data, signature.Bytes()) == nil
}

// Bytes returns a marshaled version of the PublicKey.
func (p PublicKey) Bytes() []byte {
	bytes, err := p.Point.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return bytes
}

// Base58 returns a base58 encoded version of the PublicKey.
func (p PublicKey) Base58() string {
	return base58.Encode(p.Bytes())
}

// String returns a human-readable version of the PublicKey (base58 encoded).
func (p PublicKey) String() string {
	return base58.Encode(p.Bytes())
}
