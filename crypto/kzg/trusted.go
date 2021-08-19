package kzg

import (
	"bytes"
	"encoding/binary"
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
	Suite         *bn256.Suite
	D             uint16
	Omega         kyber.Scalar  // persistent. omega: a primitive root of unity of the field. If omega==0, it is not used
	LagrangeBasis []kyber.Point // persistent. TLi = [l<i>(secret)]1
	Diff2         []kyber.Point // persistent
	// auxiliary values
	// precalc
	Domain        []kyber.Scalar // non-persistent. if omega != 0, domain_i =  omega^i, otherwise domain_i = i.
	AprimeDomainI []kyber.Scalar
	ZeroG1        kyber.Scalar
	OneG1         kyber.Scalar
}

var (
	errWrongSecret = xerrors.New("wrong secret")
	errNotROU      = xerrors.New("not a root of unity")
	errWrongROU    = xerrors.New("wrong root of unity")
)

func newTrustedSetup(suite *bn256.Suite) *TrustedSetup {
	return &TrustedSetup{Suite: suite}
}

func (sd *TrustedSetup) init(d uint16) {
	sd.D = d
	sd.Omega = sd.Suite.G1().Scalar()
	sd.LagrangeBasis = make([]kyber.Point, d)
	sd.Diff2 = make([]kyber.Point, d)
	sd.Domain = make([]kyber.Scalar, d)
	sd.AprimeDomainI = make([]kyber.Scalar, d)
	for i := range sd.Domain {
		sd.Domain[i] = sd.Suite.G1().Scalar()
	}
	for i := range sd.AprimeDomainI {
		sd.AprimeDomainI[i] = sd.Suite.G1().Scalar()
	}
	for i := range sd.LagrangeBasis {
		sd.LagrangeBasis[i] = sd.Suite.G1().Point()
		sd.Diff2[i] = sd.Suite.G2().Point()
	}
	sd.ZeroG1 = sd.Suite.G1().Scalar().Zero()
	sd.OneG1 = sd.Suite.G1().Scalar().One()
}

// TrustedSetupFromSecretPowers calculates TrustedSetup from secret and omega
// It uses powers of the omega as a domain for Lagrange basis
// Only used once after what secret must be destroyed
// The trusted setup does not contain any secret
func TrustedSetupFromSecretPowers(suite *bn256.Suite, d uint16, omega, secret kyber.Scalar) (*TrustedSetup, error) {
	ret := newTrustedSetup(suite)
	ret.init(d)
	if err := ret.generatePowers(omega, secret); err != nil {
		return nil, err
	}
	return ret, nil
}

func TrustedSetupFromSecretLinearDomain(suite *bn256.Suite, d uint16, secret kyber.Scalar) (*TrustedSetup, error) {
	ret := newTrustedSetup(suite)
	ret.init(d)
	if err := ret.generateLinearDomain(secret); err != nil {
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
	if !ret.Omega.Equal(ret.ZeroG1) {
		for i := range ret.Domain {
			powerSimple(ret.Suite, ret.Omega, i, ret.Domain[i])
			if i > 0 && ret.Domain[i].Equal(ret.OneG1) {
				return nil, errWrongROU
			}
		}
	} else {
		for i := range ret.Domain {
			ret.Domain[i].SetInt64(int64(i))
		}
	}
	for i := range ret.AprimeDomainI {
		ret.aprime(i, ret.AprimeDomainI[i])
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

// generatePowers creates a new TrustedSetup based on omega and secret
func (sd *TrustedSetup) generatePowers(omega, secret kyber.Scalar) error {
	if len(secret.String()) < 50 {
		return errWrongSecret
	}
	sd.Omega.Set(omega)
	for i := range sd.Domain {
		powerSimple(sd.Suite, sd.Omega, i, sd.Domain[i])
		if sd.Domain[i].Equal(secret) {
			return errWrongSecret
		}
		if i > 0 && sd.Domain[i].Equal(sd.OneG1) {
			return errWrongROU
		}
	}
	for i := range sd.AprimeDomainI {
		sd.aprime(i, sd.AprimeDomainI[i])
	}
	// calculate Lagrange basis: [l_i(s)]1
	for i := range sd.LagrangeBasis {
		l := sd.evalLagrangeValue(i, secret)
		sd.LagrangeBasis[i].Mul(l, nil) // [l_i(secret)]1
	}
	// calculate [secret-rou^i]2
	e2 := sd.Suite.G2().Scalar()
	for i := range sd.Diff2 {
		e2.Sub(secret, sd.Domain[i])
		sd.Diff2[i].Mul(e2, nil)
	}
	return nil
}

// generateLinearDomain creates a new TrustedSetup from secret and using 0,1,2..d-1 as a domain for Lagrange basis
func (sd *TrustedSetup) generateLinearDomain(secret kyber.Scalar) error {
	if len(secret.String()) < 50 {
		return errWrongSecret
	}
	sd.Omega.Zero()
	for i := range sd.Domain {
		sd.Domain[i].SetInt64(int64(i))
	}
	for i := range sd.AprimeDomainI {
		sd.aprime(i, sd.AprimeDomainI[i])
	}
	// calculate Lagrange basis: [l_i(s)]1
	for i := range sd.LagrangeBasis {
		l := sd.evalLagrangeValue(i, secret)
		sd.LagrangeBasis[i].Mul(l, nil) // [l_i(secret)]1
	}
	// calculate [secret-domain_i]2
	e2 := sd.Suite.G2().Scalar()
	for i := range sd.Diff2 {
		e2.Sub(secret, sd.Domain[i])
		sd.Diff2[i].Mul(e2, nil)
	}
	return nil
}

// evalLagrangeValue calculates li(X) = [prod<j=0,D-1;j!=i>((X-omega^j)/(omega^i-omega^j)]1
func (sd *TrustedSetup) evalLagrangeValue(i int, v kyber.Scalar) kyber.Scalar {
	ret := sd.Suite.G1().Scalar().One()
	numer := sd.Suite.G1().Scalar()
	denom := sd.Suite.G1().Scalar()
	elem := sd.Suite.G1().Scalar()
	for j := 0; j < int(sd.D); j++ {
		if j == i {
			continue
		}
		numer.Sub(v, sd.Domain[j])
		denom.Sub(sd.Domain[i], sd.Domain[j])
		elem.Div(numer, denom)
		ret.Mul(ret, elem)
	}
	return ret
}

// A'(omega^m)
func (sd *TrustedSetup) aprime(m int, ret kyber.Scalar) {
	e := sd.Suite.G1().Scalar()
	ret.One()
	for i := range sd.Domain {
		if i == m {
			continue
		}
		e.Sub(sd.Domain[m], sd.Domain[i])
		ret.Mul(ret, e)
	}
}

// write marshal
func (sd *TrustedSetup) write(w io.Writer) error {
	var tmp2 [2]byte
	binary.LittleEndian.PutUint16(tmp2[:], sd.D)
	if _, err := w.Write(tmp2[:]); err != nil {
		return err
	}
	if _, err := sd.Omega.MarshalTo(w); err != nil {
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
	var tmp2 [2]byte
	if _, err := r.Read(tmp2[:]); err != nil {
		return err
	}

	sd.init(binary.LittleEndian.Uint16(tmp2[:]))

	if _, err := sd.Omega.UnmarshalFrom(r); err != nil {
		return err
	}
	if !isRootOfUnity(sd.Suite, sd.Omega) {
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
