# Package KZG
Package KZG implements functions needed for _Kate-Zaverucha-Goldberg_ polynomial commitments,
also known as KZG, KZG10 and **Kate commitments**.

The KZG commitments are needed to calculate vector commitments for the implementation
of `verkle trees`, a more efficient variation of a classic Merkle trees.

See:
* [Constant-Size Commitments to Polynomials and Their Applications](https://www.iacr.org/archive/asiacrypt2010/6477178/6477178.pdf),
  the original paper

The implementation uses [DEDIS Advanced Crypto Library for Go Kyber v3](https://github.com/dedis/kyber)
and its `BN256` bilinear pairing suite as cryptographic primitives.

It follows guidelines:
* [KZG polynomial commitments](https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html) by Dankrad Feist
* [PCS multiproofs using random evaluation](https://dankradfeist.de/ethereum/2021/06/18/pcs-multiproofs.html) Dankrad Feist

This implementation uses a trusted setup completely in the Lagrange basis an all calculetions
are performed in the evaluation form of the polynomial. The coordinate form and powers of the secret on the curve
nor any FFT operations are not needed.

`D` kan be arbitrary value. The implementation assumes maximum lengths of vectors `D` = 257.
It corresponds to the 257-ary _verkle trie_.

The math of the implementation is described in this [document](https://hackmd.io/JM7BDAugQyuJgW66K-OX7A).



Some more readings:

* [Kate Commitments: A Primer](https://hackmd.io/@tompocock/Hk2A7BD6U)
