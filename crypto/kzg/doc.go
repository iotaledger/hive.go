// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package kzg implements functions needed for Kate-Zaverucha-Goldberg polynomial commitments,
// also known as KZG, KZG10 and Kate commitments.
// The KZG commitments are needed for vector commitments and the implementation
// of 'verkle' trees, a more efficient variation of a classic Merkle trees.
// See:
// - https://www.iacr.org/archive/asiacrypt2010/6477178/6477178.pdf
// - https://hackmd.io/@tompocock/Hk2A7BD6U
// The implementation uses DEDIS Advanced Crypto Library for Go Kyber v3 and its
// BN256 bilinear pairing suite as cryptographic primitives.
// The implementation assumes fixed degree of polynomials D = 16.
// It follows guidelines:
// - https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html
// - https://dankradfeist.de/ethereum/2021/06/18/pcs-multiproofs.html
// This KZG package uses proprietary structure of trusted setup.
// The trusted setup contains different values on G1 curve precomputed from the secret scalar and
// generated (net secret) primitive root of unity for the field.
package kzg
