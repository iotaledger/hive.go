package kzg

import (
	"go.dedis.ch/kyber/v3"
)

// Commit commits to vector vect[0], ...., vect[D-1]
// it is [f(s)]1 where f is polynomial  in evaluation (Lagrange) form,
// i.e. with f(rou[i]) = vect[i], i = 0..D-1
// vect[k] == nil equivalent to 0
func (sd *TrustedSetup) Commit(vect []kyber.Scalar) kyber.Point {
	ret := sd.Suite.G1().Point().Null()
	elem := sd.Suite.G1().Point()
	for i, e := range vect {
		if e == nil {
			continue
		}
		elem.Mul(e, sd.LagrangeBasis[i])
		ret.Add(ret, elem)
	}
	return ret
}

// Prove returns pi = [(f(s)-vect<index>)/(s-rou<index>)]1
// This isthe proof sent to verifier
func (sd *TrustedSetup) Prove(vect []kyber.Scalar, i int) kyber.Point {
	ret := sd.Suite.G1().Point().Null()
	e := sd.Suite.G1().Point()
	qij := sd.Suite.G1().Scalar()
	for j := range sd.OmegaPowers {
		sd.q(vect, i, j, qij)
		e.Mul(qij, sd.LagrangeBasis[j])
		ret.Add(ret, e)
	}
	return ret
}

// Prove returns pi = [(f(s)-y)/(s-rou<index>)]1
// This isthe proof sent to verifier
func (sd *TrustedSetup) ProveByValue(vect []kyber.Scalar, i int, y kyber.Scalar) kyber.Point {
	ret := sd.Suite.G1().Point().Null()
	e := sd.Suite.G1().Point()
	qij := sd.Suite.G1().Scalar()
	for j := range sd.OmegaPowers {
		sd.qValue(vect, i, j, y, qij)
		e.Mul(qij, sd.LagrangeBasis[j])
		ret.Add(ret, e)
	}
	return ret
}

func (sd *TrustedSetup) q(vect []kyber.Scalar, i, m int, ret kyber.Scalar) {
	numer := sd.Suite.G1().Scalar()
	denom := sd.Suite.G1().Scalar()
	if i != m {
		sd.diffByIndex(vect, m, i, numer)
		if numer.Equal(sd.ZeroG1) {
			ret.Zero()
			return
		}
		denom.Sub(sd.OmegaPowers[m], sd.OmegaPowers[i])
		ret.Div(numer, denom)
		return
	}
	// i == m
	ret.Zero()
	for j := range vect {
		if j == m {
			continue
		}
		sd.diffByIndex(vect, j, m, numer)
		if numer.Equal(sd.ZeroG1) {
			continue
		}
		numer.Mul(numer, sd.AprimeOmegaI[m])
		denom.Sub(sd.OmegaPowers[m], sd.OmegaPowers[j])
		denom.Mul(denom, sd.AprimeOmegaI[j])
		numer.Div(numer, denom)
		ret.Add(ret, numer)
	}
}

// used for testing
func (sd *TrustedSetup) qValue(vect []kyber.Scalar, i, m int, y kyber.Scalar, ret kyber.Scalar) {
	numer := sd.Suite.G1().Scalar()
	denom := sd.Suite.G1().Scalar()
	if i != m {
		sd.diff(vect[m], y, numer)
		if numer.Equal(sd.ZeroG1) {
			ret.Zero()
			return
		}
		denom.Sub(sd.OmegaPowers[m], sd.OmegaPowers[i])
		ret.Div(numer, denom)
		return
	}
	// i == m
	ret.Zero()
	for j := range vect {
		if j == m {
			continue
		}
		sd.diff(vect[j], y, numer)
		if numer.Equal(sd.ZeroG1) {
			continue
		}
		numer.Mul(numer, sd.AprimeOmegaI[m])
		denom.Sub(sd.OmegaPowers[m], sd.OmegaPowers[j])
		denom.Mul(denom, sd.AprimeOmegaI[j])
		numer.Div(numer, denom)
		ret.Add(ret, numer)
	}
}

func (sd *TrustedSetup) diffByIndex(vect []kyber.Scalar, i, j int, ret kyber.Scalar) {
	sd.diff(vect[i], vect[j], ret)
}

func (sd *TrustedSetup) diff(vi, vj kyber.Scalar, ret kyber.Scalar) {
	switch {
	case vi == nil && vj == nil:
		ret.Zero()
		return
	case vi != nil && vj == nil:
		ret.Set(vi)
	case vi == nil && vj != nil:
		ret.Neg(vj)
	default:
		ret.Sub(vi, vj)
	}
}

// Verify verifies KZG proof that polynomial f committed with C has f(rou<atIndex>) = v
// c is commitment to the polynomial
// pi is commitment to the value point (proof)
// value is the value of the polynomial
// adIndex is index of the root of unity where polynomial is expected to have value = v
func (sd *TrustedSetup) Verify(c, pi kyber.Point, v kyber.Scalar, atIndex int) bool {
	p1 := sd.Suite.Pair(pi, sd.Diff2[atIndex])
	e := sd.Suite.G1().Point().Mul(v, nil)
	e.Sub(c, e)
	p2 := sd.Suite.Pair(e, sd.Suite.G2().Point().Base())
	return p1.Equal(p2)
}

// VerifyVector calculates proofs and verifies all elements in the vector against commitment C
func (sd *TrustedSetup) VerifyVector(vect []kyber.Scalar, c kyber.Point) bool {
	pi := make([]kyber.Point, sd.D)
	for i := range vect {
		pi[i] = sd.Prove(vect, i)
	}
	for i := range pi {
		v := vect[i]
		if v == nil {
			v = sd.ZeroG1
		}
		if !sd.Verify(c, pi[i], v, i) {
			return false
		}
	}
	return true
}

// CommitAll return commit to the whole vector and to each of values of it
// Generate commitment to the vector and proofs to all values.
// Expensive. Usually used only in tests
func (sd *TrustedSetup) CommitAll(vect []kyber.Scalar) (kyber.Point, []kyber.Point) {
	retC := sd.Commit(vect)
	retPi := make([]kyber.Point, sd.D)
	for i := range vect {
		if vect[i] == nil {
			continue
		}
		retPi[i] = sd.Prove(vect, i)
	}
	return retC, retPi
}
