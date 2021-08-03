# Package KZG
Package KZG implements functions needed for _Kate-Zaverucha-Goldberg_ polynomial commitments,
also known as KZG, KZG10 and **Kate commitments**.

The KZG commitments are needed to calculate vector committments for the implementation
of `verkle trees`, a more efficient variation of a classic Merkle trees.

See:
* [Constant-Size Commitments to Polynomials and Their Applications](https://www.iacr.org/archive/asiacrypt2010/6477178/6477178.pdf),
  the original paper

The implementation uses [DEDIS Advanced Crypto Library for Go Kyber v3](https://github.com/dedis/kyber)
and its `BN256` bilinear pairing suite as cryptographic primitives.

The implementation assumes fixed degree of polynomials D = 16. It follows guidelines:
* [KZG polynomial commitments](https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html) by Dankrad Feist
* [PCS multiproofs using random evaluation](https://dankradfeist.de/ethereum/2021/06/18/pcs-multiproofs.html) Dankrad Feist

However, this implementation uses proprietary structure of the trusted setup.
The implemented trusted setup contains different values on G1 curve which are precomputed from the secret scalar and
generated (not secret) primitive root of unity for the field.

Some more readings:

* [Kate Commitments: A Primer](https://hackmd.io/@tompocock/Hk2A7BD6U)
