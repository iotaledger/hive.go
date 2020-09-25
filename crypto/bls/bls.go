package bls

import (
	"fmt"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

const (
	// PublicKeySize represents the length in bytes of a BLS public key.
	PublicKeySize = 128

	// SignatureSize represents the length in bytes of a BLS signature.
	SignatureSize = 64

	// PrivateKeySize represents the length in bytes of a BLS private key.
	PrivateKeySize = 32
)

func AggregateSignatures(signatures ...SignatureWithPublicKey) (signatureWithPublicKey SignatureWithPublicKey, err error) {
	if len(signatures) == 0 {
		err = fmt.Errorf("must be at least one signature to aggregate")
		return
	}
	if len(signatures) == 1 {
		signatureWithPublicKey = signatures[0]
		return
	}

	publicKeyPoints := make([]kyber.Point, len(signatures))
	signaturesBytes := make([][]byte, len(signatures))
	for i, signature := range signatures {
		publicKeyPoints[i] = signature.PublicKey().point
		signaturesBytes[i] = signature.Signature().Bytes()
	}

	mask, _ := sign.NewMask(blsSuite, publicKeyPoints, nil)
	for i := range publicKeyPoints {
		_ = mask.SetBit(i, true)
	}

	aggregatedSignature, err := bdn.AggregateSignatures(blsSuite, signaturesBytes, mask)
	if err != nil {
		return
	}
	aggregatedPublicKey, err := bdn.AggregatePublicKeys(blsSuite, mask)
	if err != nil {
		return
	}

	signatureBytes, err := aggregatedSignature.MarshalBinary()
	if err != nil {
		return
	}

	copy(signatureWithPublicKey.signature[:], signatureBytes)
	signatureWithPublicKey.publicKey.point = aggregatedPublicKey

	return
}

// blsSuite is required to perform the BLS operations.
var blsSuite = bn256.NewSuite()
