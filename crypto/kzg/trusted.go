package kzg

import (
	"bytes"
	"io"
	"io/ioutil"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"golang.org/x/xerrors"
)

// TrustedSetup is a trusted setup for KZG calculations with degree D.
// It includes roots of unity for the scalar field and values on G1 and G2 curves which are pre-calculated
// from the secret. The secret itself must be destroyed immediately after trusted setup is generated.
// The trusted setup is a public value. Normally it is stored as a public constant or file and is loaded from there.
// It is impossible to restore secret from the trusted setup
// [x]1 means a projection of scalar x to the G1 curve. [x]1 = xG, where G is the generating element
// [x]2 means a projection of scalar x to the G2 curve. [x]2 = xH, where H is the generating element
type TrustedSetup struct {
	Suite             *bn256.Suite
	RootOfUnity       kyber.Scalar    // omega: a primitive root of unity of the field.
	RootOfUnityPowers [D]kyber.Scalar // omega<i> =  omega^i. omega<0> == 1, omega<1> =  omega
	LagrangeBasis     [D]kyber.Point  // TLi = [l<i>(secret)]1
	Diff2             [D]kyber.Point
	// auxiliar precalculated values
	TA     [D][D]kyber.Scalar
	ZeroG1 kyber.Scalar
	OneG1  kyber.Scalar
	DG1    kyber.Scalar
}

var (
	errWrongSecret = xerrors.New("wrong secret")
	errNotROU      = xerrors.New("not a root of unity")
	errWrongROU    = xerrors.New("wrong root of unity")
)

func newTrustedSetup(suite *bn256.Suite) *TrustedSetup {
	ret := &TrustedSetup{
		Suite:       suite,
		RootOfUnity: suite.G1().Scalar(),
		// util
		ZeroG1: suite.G1().Scalar().Zero(),
		OneG1:  suite.G1().Scalar().One(),
		DG1:    suite.G1().Scalar().SetInt64(D),
	}
	for i := range ret.RootOfUnityPowers {
		ret.RootOfUnityPowers[i] = suite.G1().Scalar()
	}
	for i := 0; i < D; i++ {
		ret.LagrangeBasis[i] = suite.G1().Point()
		ret.Diff2[i] = suite.G2().Point()
		for j := 0; j < D; j++ {
			if i != j {
				ret.TA[i][j] = suite.G1().Scalar()
			}
		}
	}
	return ret
}

// TrustedSetupFromSecret calculates TrustedSetup from secret and rootOfUnity
// Only used once after what secret must be destroyed
// The trusted setup does not contain any secret
func TrustedSetupFromSecret(suite *bn256.Suite, rootOfUnity, secret kyber.Scalar) (*TrustedSetup, error) {
	ret := newTrustedSetup(suite)
	if err := ret.generate(rootOfUnity, secret); err != nil {
		return nil, err
	}
	if err := ret.precalcAux(); err != nil {
		return nil, err
	}
	return ret, nil
}

// TrustedSetupFromBytes unmarshals trusted setup from binary representation
func TrustedSetupFromBytes(suite *bn256.Suite, data []byte) (*TrustedSetup, error) {
	ret := newTrustedSetup(suite)
	if err := ret.read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	if err := ret.precalcAux(); err != nil {
		return nil, err
	}
	return ret, nil
}

// TrustedSetupFromFile restores trusted setup from file
func TrustedSetupFromFile(suite *bn256.Suite, fname string) (*TrustedSetup, error) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	ret, err := TrustedSetupFromBytes(suite, data)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Bytes marshals the trusted setup
func (sd *TrustedSetup) Bytes() []byte {
	var buf bytes.Buffer
	if err := sd.write(&buf); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// generate creates a new TrustedSetup based on its root of unity and provided secret
func (sd *TrustedSetup) generate(rootOfUnity, secret kyber.Scalar) error {
	if len(secret.String()) < 50 {
		return errWrongSecret
	}
	sd.RootOfUnity.Set(rootOfUnity)
	if err := sd.precalcRootsOfUnity(secret); err != nil {
		return err
	}
	// calculate Lagrange basis: [l_i(s)]1
	for i := range sd.LagrangeBasis {
		l := sd.evalLagrangeValue(i, secret)
		sd.LagrangeBasis[i].Mul(l, nil) // [l_i(secret)]1
	}
	// calculate [secret-rou^i]2
	e2 := sd.Suite.G2().Scalar()
	for i := range sd.Diff2 {
		e2.Sub(secret, sd.RootOfUnityPowers[i])
		sd.Diff2[i].Mul(e2, nil)
	}
	return nil
}

// precalcRootsOfUnity calculates powers up to D-1 of roots of unity
func (sd *TrustedSetup) precalcRootsOfUnity(secret kyber.Scalar) error {
	for i := range sd.RootOfUnityPowers {
		powerSimple(sd.Suite, sd.RootOfUnity, i, sd.RootOfUnityPowers[i])
		if sd.RootOfUnityPowers[i].Equal(secret) {
			return errWrongSecret
		}
		if i > 0 && sd.RootOfUnityPowers[i].Equal(sd.OneG1) {
			return errWrongROU
		}
	}
	return nil
}

func (sd *TrustedSetup) precalcAux() error {
	e := sd.Suite.G1().Scalar()
	for m := range sd.TA {
		for j := range sd.TA {
			if m == j {
				continue
			}
			sd.TA[m][j].Div(sd.aprime(m), sd.aprime(j))
			e.Sub(sd.RootOfUnityPowers[m], sd.RootOfUnityPowers[j])
			sd.TA[m][j].Div(sd.TA[m][j], e)
		}
	}
	return nil
}

// aprime return A'(\omega^k)
func (sd *TrustedSetup) aprime(k int) kyber.Scalar {
	ret := sd.Suite.G1().Scalar().One()
	e := sd.Suite.G1().Scalar()
	for j := range sd.RootOfUnityPowers {
		if k == j {
			continue
		}
		e.Sub(sd.RootOfUnityPowers[k], sd.RootOfUnityPowers[j])
		ret.Mul(ret, e)
	}
	return ret
}

// evalLagrangeValue calculates li(X) = [prod<j=0,D-1;j!=i>((X-omega^j)/(omega^i-omega^j)]1
func (sd *TrustedSetup) evalLagrangeValue(i int, v kyber.Scalar) kyber.Scalar {
	ret := sd.Suite.G1().Scalar().One()
	numer := sd.Suite.G1().Scalar()
	denom := sd.Suite.G1().Scalar()
	elem := sd.Suite.G1().Scalar()
	for j := 0; j < D; j++ {
		if j == i {
			continue
		}
		numer.Sub(v, sd.RootOfUnityPowers[j])
		denom.Sub(sd.RootOfUnityPowers[i], sd.RootOfUnityPowers[j])
		elem.Div(numer, denom)
		ret.Mul(ret, elem)
	}
	return ret
}

// write marshal
func (sd *TrustedSetup) write(w io.Writer) error {
	if _, err := sd.RootOfUnity.MarshalTo(w); err != nil {
		return err
	}
	for i := range sd.LagrangeBasis {
		if _, err := sd.LagrangeBasis[i].MarshalTo(w); err != nil {
			return err
		}
	}
	for i := range sd.Diff2 {
		if _, err := sd.Diff2[i].MarshalTo(w); err != nil {
			return err
		}
	}
	return nil
}

// read unmarshal
func (sd *TrustedSetup) read(r io.Reader) error {
	if _, err := sd.RootOfUnity.UnmarshalFrom(r); err != nil {
		return err
	}
	if !isRootOfUnity(sd.Suite, sd.RootOfUnity) {
		return errNotROU
	}
	for i := range sd.LagrangeBasis {
		if _, err := sd.LagrangeBasis[i].UnmarshalFrom(r); err != nil {
			return err
		}
	}
	for i := range sd.Diff2 {
		if _, err := sd.Diff2[i].UnmarshalFrom(r); err != nil {
			return err
		}
	}
	return nil
}
