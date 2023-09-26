package bitmask

type BitMask byte

// SetBit sets the bit at the given position.
func (bitmask BitMask) SetBit(pos uint) BitMask {
	return bitmask.SetBits(1 << pos)
}

// SetBits sets the bits in the given bitmask.
func (bitmask BitMask) SetBits(bits BitMask) BitMask {
	return bitmask | bits
}

// ClearBit clears the bit at the given position.
func (bitmask BitMask) ClearBit(pos uint) BitMask {
	return bitmask.ClearBits(1 << pos)
}

// ClearBits clears the bits in the given bitmask.
func (bitmask BitMask) ClearBits(bits BitMask) BitMask {
	return bitmask & ^bits
}

// HasBit checks whether the bit at the given position is set.
func (bitmask BitMask) HasBit(pos uint) bool {
	return bitmask.HasBits(1 << pos)
}

// HasBits checks whether the bits in the given bitmask are set.
func (bitmask BitMask) HasBits(bits BitMask) bool {
	return bitmask&(bits) > 0
}

// ModifyBit sets or clears the bit at the given position, given the supplied state bool.
func (bitmask BitMask) ModifyBit(pos uint, state bool) BitMask {
	if state {
		return bitmask.SetBit(pos)
	}

	return bitmask.ClearBit(pos)
}
