// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package kzg implements functions needed for Kate-Zaverucha-Goldberg polynomial commitments,
// also know as KZG, KZG10 and Kate commitments.
// The KZG commitments are needed for vector commitments and the implementation
// of 'verkle' trees, a more efficient variation of a classic Merkle trees.
// See:
// - https://www.iacr.org/archive/asiacrypt2010/6477178/6477178.pdf
// - https://hackmd.io/@tompocock/Hk2A7BD6U
// The implementation uses DEDIS Advanced Crypto Library for Go Kyber v.3 for BN256 bilinear pairing suite for
// cryptographic primitives.
// The implementation assumes fixed degree of polynomials D = 16.
// It follows guidelines:
// - https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html
// - https://dankradfeist.de/ethereum/2021/06/18/pcs-multiproofs.html
package kzg
