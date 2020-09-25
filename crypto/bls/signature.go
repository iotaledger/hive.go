package bls

import (
	"github.com/iotaledger/hive.go/byteutils"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

// region Signature ////////////////////////////////////////////////////////////////////////////////////////////////////

// Signature is the type of a raw BLS signature.
type Signature [SignatureSize]byte

// SignatureFromBytes unmarshals a Signature from a sequence of bytes.
func SignatureFromBytes(bytes []byte) (signature Signature, consumedBytes int, err error) {
	marshalUtil := marshalutil.New(bytes)
	if signature, err = SignatureFromMarshalUtil(marshalUtil); err != nil {
		err = xerrors.Errorf("failed to parse Signature from MarshalUtil: %w", err)
		return
	}
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// SignatureFromBase58EncodedString creates a Signature from a base58 encoded string.
func SignatureFromBase58EncodedString(base58EncodedString string) (signature Signature, err error) {
	bytes, err := base58.Decode(base58EncodedString)
	if err != nil {
		err = xerrors.Errorf("error while decoding base58 encoded Signature (%v): %w", err, ErrBase58DecodeFailed)
		return
	}

	if signature, _, err = SignatureFromBytes(bytes); err != nil {
		err = xerrors.Errorf("failed to parse Signature from bytes: %w", err)
		return
	}

	return
}

// SignatureFromMarshalUtil unmarshals a Signature using a MarshalUtil (for easier unmarshaling).
func SignatureFromMarshalUtil(marshalUtil *marshalutil.MarshalUtil) (signature Signature, err error) {
	signatureBytes, err := marshalUtil.ReadBytes(SignatureSize)
	if err != nil {
		err = xerrors.Errorf("failed to read signature bytes (%v): %w", err, ErrParseBytesFailed)
		return
	}
	copy(signature[:], signatureBytes)

	return
}

// Bytes returns a marshaled version of the Signature.
func (s Signature) Bytes() []byte {
	return s[:]
}

// Base58 returns a base58 encoded version of the Signature.
func (s Signature) Base58() string {
	return base58.Encode(s.Bytes())
}

// String returns a human readable version of the signature.
func (s Signature) String() string {
	return s.Base58()
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region SignatureWithPublicKey ///////////////////////////////////////////////////////////////////////////////////////

// SignatureWithPublicKey is a combination of a PublicKey and a Signature that is required to perform operations like
// Signature- and PublicKey-aggregations.
type SignatureWithPublicKey struct {
	PublicKey PublicKey
	Signature Signature
}

// NewSignatureWithPublicKey is the constructor for SignatureWithPublicKey objects.
func NewSignatureWithPublicKey(publicKey PublicKey, signature Signature) SignatureWithPublicKey {
	return SignatureWithPublicKey{
		PublicKey: publicKey,
		Signature: signature,
	}
}

// IsValid returns true if the signature is correct for the given data.
func (s SignatureWithPublicKey) IsValid(data []byte) bool {
	return s.PublicKey.SignatureValid(data, s.Signature)
}

// Bytes returns the signature in bytes.
func (s SignatureWithPublicKey) Bytes() []byte {
	return byteutils.ConcatBytes(s.PublicKey.Bytes(), s.Signature.Bytes())
}

// String returns a human readable version of the SignatureWithPublicKey (base58 encoded).
func (s SignatureWithPublicKey) String() string {
	return base58.Encode(s.Bytes())
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
