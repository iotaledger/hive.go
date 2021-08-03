package kzg

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/util/random"
	"math/big"
)

// powerSimple x^n, linear multiplication
func powerSimple(suite *bn256.Suite, x kyber.Scalar, n int, setTo ...kyber.Scalar) kyber.Scalar {
	var ret kyber.Scalar
	if len(setTo) > 0 {
		ret = setTo[0]
	} else {
		ret = suite.G1().Scalar()
	}
	ret.One()
	for i := 0; i < n; i++ {
		ret.Mul(ret, x)
	}
	return ret
}

// powerBig x^n on the field where n is any big.Int
// apparently it uses exponentiation by squaring: https://en.wikipedia.org/wiki/Exponentiation_by_squaring
func powerBig(suite *bn256.Suite, x kyber.Scalar, n *big.Int) kyber.Scalar {
	if n.Cmp(big0) == 0 {
		return suite.G1().Scalar().One()
	}
	ret := suite.G1().Scalar().Set(x)
	next := new(big.Int)
	remain := new(big.Int)
	pow := new(big.Int).Set(big1)
	for {
		if pow.Cmp(n) == 0 {
			return ret
		}
		next.Mul(pow, big2)
		if next.Cmp(n) > 0 {
			remain.Sub(n, pow)
			resRemain := powerBig(suite, x, remain)
			ret.Mul(ret, resRemain)
			return ret
		}
		pow.Set(next)
		ret.Mul(ret, ret)
	}
}

// power2n is x^(2^n)
func power2n(suite *bn256.Suite, x kyber.Scalar, n uint) kyber.Scalar {
	ret := suite.G1().Scalar().Set(x)
	for i := uint(0); i < n; i++ {
		ret.Mul(ret, ret)
	}
	return ret
}

// isRootOfUnity checks if scalar is a root of unity
func isRootOfUnity(suite *bn256.Suite, rootOfUnity kyber.Scalar) bool {
	return power2n(suite, rootOfUnity, LOGD).Equal(suite.G1().Scalar().One())
}

// generateRootOfUnity generates random scalar s and returns s^((fieldOrder-1)/D)
// It is a root of unity with property rou^D == 1, however it is not necessarily a primitive root of unity.
// The primitive root of unity must satisfy rou^N != 1 for any N=1..D-1
func generateRootOfUnity(suite *bn256.Suite) kyber.Scalar {
	for {
		s := suite.G1().Scalar().Pick(random.New())
		ret := powerBig(suite, s, orderMinus1DivD)
		if len(ret.String()) >= 64 {
			return ret
		}
	}
}

// GenRootOfUnityPrimitive generates random roots of unity until all its powers are long enough thus excluding also 1
func GenRootOfUnityPrimitive(suite *bn256.Suite) (kyber.Scalar, *[D]kyber.Scalar) {
	repeat := true
	var rou kyber.Scalar
	retPowers := new([D]kyber.Scalar)

	for repeat {
		repeat = false
		rou = generateRootOfUnity(suite)
		for i := range retPowers {
			retPowers[i] = powerSimple(suite, rou, i)
			if i > 0 && len(retPowers[i].String()) < 50 {
				repeat = true
				break
			}
		}
	}
	return rou, retPowers
}
