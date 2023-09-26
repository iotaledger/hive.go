package mana

import (
	"fmt"
	"sort"

	"github.com/izuc/zipp.foundation/crypto/identity"
)

// Func type is an adapter to allow the use of ordinary functions as Mana.
type Func func(identity *identity.Identity) uint64

// Eval calls f(p).
func (f Func) Eval(identity *identity.Identity) uint64 { return f(identity) }

// Identity is a map between mana value and identities with that mana value.
type Identity map[uint64][]*identity.Identity

// NewIdentity returns the mana-identities map.
func NewIdentity(identities []*identity.Identity, f Func) Identity {
	manaRank := make(Identity)
	for _, identity := range identities {
		mana := f.Eval(identity)
		manaRank[mana] = append(manaRank[mana], identity)
	}

	return manaRank
}

// RankByVariableRange ranks by a variable range based on ro.
func RankByVariableRange(f Func, target *identity.Identity, identities []*identity.Identity, r int, ro float64) []*identity.Identity {
	manaRank := NewIdentity(identities, f)
	targetMana := f(target)
	if _, exist := manaRank[targetMana]; !exist {
		manaRank[targetMana] = []*identity.Identity{target}
	}

	if len(manaRank) < 2*r {
		return identities
	}

	upperSet := []*identity.Identity{}
	lowerSet := []*identity.Identity{}
	totalSet := []*identity.Identity{}

	if targetMana != 0 {
		for mana, identities := range manaRank {
			switch mana > targetMana {
			case true:
				if float64(mana)/float64(targetMana) < ro {
					upperSet = append(upperSet, identities...)
				}
			case false:
				if mana == 0 || mana == targetMana {
					break
				}
				if float64(targetMana)/float64(mana) < ro {
					lowerSet = append(lowerSet, identities...)
				}
			}
		}
	}

	if len(upperSet) < r || len(lowerSet) < r {

		// sort manaRank
		keys := make([]uint64, len(manaRank))
		i := 0
		for mana := range manaRank {
			keys[i] = mana
			i++
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

		// find target position
		targetIndex := findIndex(targetMana, keys)

		if len(lowerSet) < r {
			// clear lowerSet
			lowerSet = []*identity.Identity{}

			// add to totalSet the r largest mana holders less than targetMana
			for _, k := range filter(targetIndex, keys, r, down) {
				totalSet = append(totalSet, manaRank[k]...)
			}
		}
		if len(upperSet) < r {
			// clear upperSet
			upperSet = []*identity.Identity{}

			// add to totalSet the r smallest mana holders greater than targetMana
			for _, k := range filter(targetIndex, keys, r, up) {
				totalSet = append(totalSet, manaRank[k]...)
			}
		}
	}

	totalSet = append(totalSet, lowerSet...)

	// include my rank
	for _, identity := range manaRank[targetMana] {
		if identity != target {
			totalSet = append(totalSet, identity)
		}
	}

	totalSet = append(totalSet, upperSet...)

	return totalSet

}

// RankByFixedRange ranks by a fixed range r.
func RankByFixedRange(f Func, target *identity.Identity, identities []*identity.Identity, r int) []*identity.Identity {
	manaRank := NewIdentity(identities, f)
	targetMana := f(target)
	if _, exist := manaRank[targetMana]; !exist {
		manaRank[targetMana] = []*identity.Identity{target}
	}

	if len(manaRank) < 2*r {
		fmt.Println("returning full identities", len(manaRank))

		return identities
	}

	// sort manaRank
	keys := make([]uint64, len(manaRank))
	i := 0
	for mana := range manaRank {
		keys[i] = mana
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	// find target position
	targetIndex := findIndex(targetMana, keys)

	// compute upper set
	upperSet := filter(targetIndex, keys, r, up)

	// compute lower set
	lowerSet := filter(targetIndex, keys, r, down)

	// return (upper+lower) set
	// include lower ranks
	totalSet := []*identity.Identity{}
	for _, k := range lowerSet {
		totalSet = append(totalSet, manaRank[k]...)
	}

	// include my rank
	for _, identity := range manaRank[targetMana] {
		if identity != target {
			totalSet = append(totalSet, identity)
		}
	}

	// include upper ranks
	for _, k := range upperSet {
		totalSet = append(totalSet, manaRank[k]...)
	}

	return totalSet
}

// RankByThreshold ranks by mana threshold.
func RankByThreshold(f Func, target *identity.Identity, identities []*identity.Identity, threshold float64) []*identity.Identity {
	var upperCounter, lowerCounter int
	targetMana := f(target)
	totalMana := Total(f, identities) - targetMana
	var sumMana uint64
	thresholdMana := uint64(float64(totalMana) * threshold)

	totalSet := []*identity.Identity{}

	manaRank := NewIdentity(identities, f)
	if _, exist := manaRank[targetMana]; !exist {
		manaRank[targetMana] = []*identity.Identity{target}
	}

	// sort manaRank
	keys := make([]uint64, len(manaRank))
	i := 0
	for mana := range manaRank {
		keys[i] = mana
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	// find target position
	targetIndex := findIndex(targetMana, keys)

	lowerCounter = targetIndex - 1
	upperCounter = targetIndex + 1

	for sumMana < thresholdMana && (upperCounter < len(keys) || lowerCounter >= 0) {
		if lowerCounter >= 0 {
			sumMana += keys[lowerCounter]
			totalSet = append(totalSet, manaRank[keys[lowerCounter]]...)
			lowerCounter--
		}
		if upperCounter < len(keys) {
			sumMana += keys[upperCounter]
			totalSet = append(totalSet, manaRank[keys[upperCounter]]...)
			upperCounter++
		}
	}

	// include my rank
	for _, identity := range manaRank[targetMana] {
		if identity != target {
			totalSet = append(totalSet, identity)
		}
	}

	return totalSet
}

func findIndex(target uint64, set []uint64) int {
	for i, v := range set {
		if v == target {
			return i
		}
	}

	return -1
}

const (
	down = -1
	up   = +1
)

func filter(target int, set []uint64, r int, mode int) []uint64 {
	switch mode {
	case down:
		start := target - r
		if start < 0 {
			start = 0
		}

		return set[start:target]
	case up:
		end := target + r + 1
		if end > len(set) {
			end = len(set)
		}
		if target+1 >= len(set) {
			return []uint64{}
		}

		return set[target+1 : end]
	default:
		return nil
	}
}

// Total returns the total mana of the given identities.
func Total(f Func, identities []*identity.Identity) (total uint64) {
	for _, identity := range identities {
		total += f(identity)
	}

	return total
}
