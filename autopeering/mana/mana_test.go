package mana

import (
	"fmt"
	"log"
	"math"
	"testing"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/identity"
	"github.com/stretchr/testify/require"
)

// 100
// 20
// 10 10
// 4 4 4 4
// 3 3 3  3
// 2 2 2 2 2 2
// 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1

func newTestIdentity(name string) *identity.Identity {
	key := ed25519.PublicKey{}
	copy(key[:], name)
	return identity.New(key)
}

func newTestIdentities(n int) (identities []*identity.Identity) {
	identities = make([]*identity.Identity, n)
	for i := 0; i < n; i++ {
		identities[i] = newTestIdentity(fmt.Sprintf("%d", i))
	}
	return
}

var testIdentityMana map[*identity.Identity]uint64

func newTestMana(identities []*identity.Identity) (m map[*identity.Identity]uint64) {
	m = make(map[*identity.Identity]uint64, len(identities))
	for i, p := range identities {
		m[p] = uint64(i)
	}
	return m
}

func newZipfMana(identities []*identity.Identity, zipf float64) (m map[*identity.Identity]uint64) {
	m = make(map[*identity.Identity]uint64, len(identities))
	scalingFactor := math.Pow(10, 10)
	for i, p := range identities {
		m[p] = uint64(math.Pow(float64(i+1), -zipf) * scalingFactor)
		//log.Println(m[p])
	}
	return m
}

type sameManaFunc Func

var sameMana sameManaFunc

func (f sameManaFunc) Eval(identity *identity.Identity) uint64 {
	return 1
}

type manaFunc Func

var mana manaFunc

func (f manaFunc) Eval(identity *identity.Identity) uint64 {
	return testIdentityMana[identity]
}

func stringID(identities []*identity.Identity) (output []string) {
	for _, item := range identities {
		output = append(output, fmt.Sprintf(item.ID().String()))
	}
	return
}

func TestZipfMana(t *testing.T) {
	IDs := newTestIdentities(100)
	testIdentityMana = newZipfMana(IDs, 0.82)
	totalMana := Total(mana.Eval, IDs)
	log.Println("Total Mana:", totalMana)

	for _, id := range IDs {
		fmt.Printf("%s - %.4f\n", id.ID().String(), float64(mana.Eval(id))/float64(totalMana))
	}
}

func TestAsymmetry(t *testing.T) {
	IDs := newTestIdentities(1000)
	testIdentityMana = newZipfMana(IDs, 0.82)
	totalMana := Total(mana.Eval, IDs)

	R := 10
	ro := 1.2
	threshold := 1. / 3.

	// largest mana holder
	log.Println("largest mana holder ")

	target := IDs[0]

	set := RankByFixedRange(mana.Eval, target, IDs, R)
	log.Println("RankByFixedRange - len:", len(set))
	occurance := 0
	for _, identity := range set {
		occurance += asymmetryCheck(target, RankByFixedRange(mana.Eval, identity, IDs, R))
	}
	log.Println("RankByFixedRange - asymmetry:", occurance)
	log.Printf("RankByFixedRange - mana: %.2f %% \n", float64(Total(mana.Eval, set))/float64(totalMana-mana.Eval(target))*100)

	set = RankByVariableRange(mana.Eval, target, IDs, R, ro)
	log.Println("RankByVariableRange - len:", len(set))
	occurance = 0
	for _, identity := range set {
		occurance += asymmetryCheck(target, RankByVariableRange(mana.Eval, identity, IDs, R, ro))
	}
	log.Println("RankByVariableRange - asymmetry:", occurance)
	log.Printf("RankByVariableRange - mana: %.2f %% \n", float64(Total(mana.Eval, set))/float64(totalMana-mana.Eval(target))*100)

	set = RankByThreshold(mana.Eval, target, IDs, threshold)
	log.Println("RankByThreshold - len:", len(set))
	occurance = 0
	for _, identity := range set {
		occurance += asymmetryCheck(target, RankByThreshold(mana.Eval, identity, IDs, threshold))
	}
	log.Println("RankByThreshold - asymmetry:", occurance)
	log.Printf("RankByThreshold - mana: %.2f %% \n", float64(Total(mana.Eval, set))/float64(totalMana-mana.Eval(target))*100)

	// smallest mana holder
	log.Println("smallest mana holder ")

	target = IDs[len(IDs)-1]

	set = RankByFixedRange(mana.Eval, target, IDs, R)
	log.Println("RankByFixedRange - len:", len(set))
	occurance = 0
	for _, identity := range set {
		occurance += asymmetryCheck(target, RankByFixedRange(mana.Eval, identity, IDs, R))
	}
	log.Println("RankByFixedRange - asymmetry:", occurance)
	log.Printf("RankByFixedRange - mana: %.2f %% \n", float64(Total(mana.Eval, set))/float64(totalMana-mana.Eval(target))*100)

	set = RankByVariableRange(mana.Eval, target, IDs, R, ro)
	log.Println("RankByVariableRange - len:", len(set))
	occurance = 0
	for _, identity := range set {
		occurance += asymmetryCheck(target, RankByVariableRange(mana.Eval, identity, IDs, R, ro))
	}
	log.Println("RankByVariableRange - asymmetry:", occurance)
	log.Printf("RankByVariableRange - mana: %.2f %% \n", float64(Total(mana.Eval, set))/float64(totalMana-mana.Eval(target))*100)

	set = RankByThreshold(mana.Eval, target, IDs, threshold)
	log.Println("RankByThreshold - len:", len(set))
	occurance = 0
	for _, identity := range set {
		occurance += asymmetryCheck(target, RankByThreshold(mana.Eval, identity, IDs, threshold))
	}
	log.Println("RankByThreshold - asymmetry:", occurance)
	log.Printf("RankByThreshold - mana: %.2f %% \n", float64(Total(mana.Eval, set))/float64(totalMana-mana.Eval(target))*100)

	// middle mana holder
	log.Println("middle mana holder ")

	target = IDs[len(IDs)/2]

	set = RankByFixedRange(mana.Eval, target, IDs, R)
	log.Println("RankByFixedRange - len:", len(set))
	occurance = 0
	for _, identity := range set {
		occurance += asymmetryCheck(target, RankByFixedRange(mana.Eval, identity, IDs, R))
	}
	log.Println("RankByFixedRange - asymmetry:", occurance)
	log.Printf("RankByFixedRange - mana: %.2f %% \n", float64(Total(mana.Eval, set))/float64(totalMana-mana.Eval(target))*100)

	set = RankByVariableRange(mana.Eval, target, IDs, R, ro)
	log.Println("RankByVariableRange - len:", len(set))
	occurance = 0
	for _, identity := range set {
		occurance += asymmetryCheck(target, RankByVariableRange(mana.Eval, identity, IDs, R, ro))
	}
	log.Println("RankByVariableRange - asymmetry:", occurance)
	log.Printf("RankByVariableRange - mana: %.2f %% \n", float64(Total(mana.Eval, set))/float64(totalMana-mana.Eval(target))*100)

	set = RankByThreshold(mana.Eval, target, IDs, threshold)
	log.Println("RankByThreshold - len:", len(set))
	occurance = 0
	for _, identity := range set {
		occurance += asymmetryCheck(target, RankByThreshold(mana.Eval, identity, IDs, threshold))
	}
	log.Println("RankByThreshold - asymmetry:", occurance)
	log.Printf("RankByThreshold - mana: %.2f %% \n", float64(Total(mana.Eval, set))/float64(totalMana-mana.Eval(target))*100)
}

func asymmetryCheck(target *identity.Identity, identities []*identity.Identity) int {
	for _, identity := range identities {
		if identity == target {
			return 0
		}
	}
	return 1
}

func TestRankByFixedRange(t *testing.T) {
	IDs := newTestIdentities(10)
	// 0 1 2 3 4 5 6 7 8 9
	testIdentityMana = newTestMana(IDs)
	R := 2

	set := RankByFixedRange(mana.Eval, IDs[0], IDs, R)
	expectedSet := IDs[1:3]
	require.ElementsMatch(t, expectedSet, set)

	set = RankByFixedRange(mana.Eval, IDs[len(IDs)-1], IDs, R)
	expectedSet = IDs[len(IDs)-1-R : len(IDs)-1]
	require.ElementsMatch(t, expectedSet, set)

	set = RankByFixedRange(mana.Eval, IDs[len(IDs)/2], IDs, R)
	expectedSet = IDs[(len(IDs)/2)-R : (len(IDs) / 2)]
	expectedSet = append(expectedSet, IDs[(len(IDs)/2)+1:(len(IDs)/2)+R+1]...)
	require.ElementsMatch(t, expectedSet, set)
}

func TestRankByVariableRange(t *testing.T) {
	IDs := newTestIdentities(10)
	// 0 1 2 3 4 5 6 7 8 9
	testIdentityMana = newTestMana(IDs)
	R := 2
	ro := 1.51

	set := RankByVariableRange(mana.Eval, IDs[0], IDs, R, ro)
	expectedSet := IDs[1:3]
	require.ElementsMatch(t, expectedSet, set)

	set = RankByVariableRange(mana.Eval, IDs[len(IDs)-1], IDs, R, ro)
	expectedSet = IDs[len(IDs)-2-R : len(IDs)-1]
	require.ElementsMatch(t, expectedSet, set)

	set = RankByVariableRange(mana.Eval, IDs[len(IDs)/2], IDs, R, ro)
	expectedSet = IDs[(len(IDs)/2)-R : (len(IDs) / 2)]
	expectedSet = append(expectedSet, IDs[(len(IDs)/2)+1:(len(IDs)/2)+R+1]...)
	require.ElementsMatch(t, expectedSet, set)
}

func TestRankByThreshold(t *testing.T) {
	IDs := newTestIdentities(10)
	testIdentityMana = newTestMana(IDs)
	threshold := 1. / 3.

	expectedMana := uint64(float64(Total(mana.Eval, IDs)) * threshold)

	set := RankByThreshold(mana.Eval, IDs[0], IDs, threshold)
	actualMana := Total(mana.Eval, set)
	require.GreaterOrEqual(t, actualMana, expectedMana)

	set = RankByThreshold(mana.Eval, IDs[len(IDs)-1], IDs, threshold)
	actualMana = Total(mana.Eval, set)
	require.GreaterOrEqual(t, actualMana, expectedMana)

	set = RankByThreshold(mana.Eval, IDs[len(IDs)/2], IDs, threshold)
	actualMana = Total(mana.Eval, set)
	require.GreaterOrEqual(t, actualMana, expectedMana)
}
