package bls

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"golang.org/x/xerrors"
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

// AggregateSignatures aggregates multiple SignatureWithPublicKey objects into a single SignatureWithPublicKey.
func AggregateSignatures(signaturesWithPublicKey ...SignatureWithPublicKey) (aggregatedSignature SignatureWithPublicKey, err error) {
	if len(signaturesWithPublicKey) == 0 {
		err = xerrors.Errorf("not enough signatures to aggregate: %w", ErrInvalidArgument)
		return
	}

	if len(signaturesWithPublicKey) == 1 {
		aggregatedSignature = signaturesWithPublicKey[0]
		return
	}

	publicKeyPoints := make([]kyber.Point, len(signaturesWithPublicKey))
	signaturesBytes := make([][]byte, len(signaturesWithPublicKey))
	for i, signatureWithPublicKey := range signaturesWithPublicKey {
		publicKeyPoints[i] = signatureWithPublicKey.PublicKey.Point
		signaturesBytes[i] = signatureWithPublicKey.Signature.Bytes()
	}

	mask, err := sign.NewMask(blsSuite, publicKeyPoints, nil)
	if err != nil {
		err = xerrors.Errorf("failed to create mask (%v): %w", err, ErrBLSFailed)
		return
	}
	for i := range publicKeyPoints {
		_ = mask.SetBit(i, true)
	}

	rawAggregatedSignature, err := bdn.AggregateSignatures(blsSuite, signaturesBytes, mask)
	if err != nil {
		err = xerrors.Errorf("failed to aggregate Signatures (%v): %w", err, ErrBLSFailed)
		return
	}
	signatureBytes, err := rawAggregatedSignature.MarshalBinary()
	if err != nil {
		err = xerrors.Errorf("failed to marshal aggregated Signature (%v): %w", err, ErrBLSFailed)
		return
	}
	copy(aggregatedSignature.Signature[:], signatureBytes)

	aggregatedSignature.PublicKey.Point, err = bdn.AggregatePublicKeys(blsSuite, mask)
	if err != nil {
		err = xerrors.Errorf("failed to aggregate PublicKeys (%v): %w", err, ErrBLSFailed)
		return
	}

	return
}
