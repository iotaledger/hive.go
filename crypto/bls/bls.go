package bls

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/util/random"

	"github.com/iotaledger/hive.go/crypto"
	"github.com/iotaledger/hive.go/ierrors"
)

const (
	// PublicKeySize represents the length in bytes of a BLS public key.
	PublicKeySize = 128

	// SignatureSize represents the length in bytes of a BLS signature.
	SignatureSize = 64

	// PrivateKeySize represents the length in bytes of a BLS private key.
	PrivateKeySize = 32
)

// blsSuite is required to perform the BLS operations of the 3rd party library.
var blsSuite = bn256.NewSuite()

// randomness contains a secure source of randomness that is used by BLS.
var randomness = random.New(crypto.Randomness)

// AggregateSignatures aggregates multiple SignatureWithPublicKey objects into a single SignatureWithPublicKey.
func AggregateSignatures(signaturesWithPublicKey ...SignatureWithPublicKey) (SignatureWithPublicKey, error) {
	if len(signaturesWithPublicKey) == 0 {
		return SignatureWithPublicKey{}, ierrors.Wrap(ErrInvalidArgument, "not enough signatures to aggregate")
	}

	if len(signaturesWithPublicKey) == 1 {
		return signaturesWithPublicKey[0], nil
	}

	publicKeyPoints := make([]kyber.Point, len(signaturesWithPublicKey))
	signaturesBytes := make([][]byte, len(signaturesWithPublicKey))
	for i, signatureWithPublicKey := range signaturesWithPublicKey {
		publicKeyPoints[i] = signatureWithPublicKey.PublicKey.Point
		signaturesBytes[i] = signatureWithPublicKey.Signature.Bytes()
	}

	mask, err := sign.NewMask(blsSuite, publicKeyPoints, nil)
	if err != nil {
		return SignatureWithPublicKey{}, ierrors.Wrapf(ErrBLSFailed, "failed to create mask: %w", err)
	}
	for i := range publicKeyPoints {
		_ = mask.SetBit(i, true)
	}

	rawAggregatedSignature, err := bdn.AggregateSignatures(blsSuite, signaturesBytes, mask)
	if err != nil {
		return SignatureWithPublicKey{}, ierrors.Wrapf(ErrBLSFailed, "failed to aggregate Signatures: %w", err)
	}
	signatureBytes, err := rawAggregatedSignature.MarshalBinary()
	if err != nil {
		return SignatureWithPublicKey{}, ierrors.Wrapf(ErrBLSFailed, "failed to marshal aggregated Signature: %w", err)
	}

	aggregatedSignature := SignatureWithPublicKey{}
	copy(aggregatedSignature.Signature[:], signatureBytes)

	aggregatedSignature.PublicKey.Point, err = bdn.AggregatePublicKeys(blsSuite, mask)
	if err != nil {
		return SignatureWithPublicKey{}, ierrors.Wrapf(ErrBLSFailed, "failed to aggregate PublicKeys: %w", err)
	}

	return aggregatedSignature, nil
}
