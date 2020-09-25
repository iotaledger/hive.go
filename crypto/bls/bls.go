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

// AggregateSignatures aggregates multiple SignatureWithPublicKey objects into a single SignatureWithPublicKey.
func AggregateSignatures(signaturesWithPublicKey ...SignatureWithPublicKey) (signatureWithPublicKey SignatureWithPublicKey, err error) {
	if len(signaturesWithPublicKey) == 0 {
		err = xerrors.Errorf("not enough signatures to aggregate: %w", ErrInvalidArgument)
		return
	}

	if len(signaturesWithPublicKey) == 1 {
		signatureWithPublicKey = signaturesWithPublicKey[0]
		return
	}

	publicKeyPoints := make([]kyber.Point, len(signaturesWithPublicKey))
	signaturesBytes := make([][]byte, len(signaturesWithPublicKey))
	for i, sigWithPublicKey := range signaturesWithPublicKey {
		publicKeyPoints[i] = sigWithPublicKey.PublicKey.Point
		signaturesBytes[i] = sigWithPublicKey.Signature.Bytes()
	}

	mask, err := sign.NewMask(blsSuite, publicKeyPoints, nil)
	if err != nil {
		err = xerrors.Errorf("failed to create mask (%v): %w", err, ErrBLSFailed)
		return
	}
	for i := range publicKeyPoints {
		_ = mask.SetBit(i, true)
	}

	aggregatedSignature, err := bdn.AggregateSignatures(blsSuite, signaturesBytes, mask)
	if err != nil {
		err = xerrors.Errorf("failed to aggregate Signatures (%v): %w", err, ErrBLSFailed)
		return
	}
	aggregatedPublicKey, err := bdn.AggregatePublicKeys(blsSuite, mask)
	if err != nil {
		err = xerrors.Errorf("failed to aggregate PublicKeys (%v): %w", err, ErrBLSFailed)
		return
	}

	signatureBytes, err := aggregatedSignature.MarshalBinary()
	if err != nil {
		err = xerrors.Errorf("failed to marshal aggregated Signature (%v): %w", err, ErrBLSFailed)
		return
	}

	copy(signatureWithPublicKey.Signature[:], signatureBytes)
	signatureWithPublicKey.PublicKey.Point = aggregatedPublicKey

	return
}

// blsSuite is required to perform the BLS operations.
var blsSuite = bn256.NewSuite()
