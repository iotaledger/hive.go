package bitmask

type BitMask byte

// SetBit sets the bit at the given position.
func (bitmask BitMask) SetBit(pos uint) BitMask {
	return bitmask.SetFlag(1 << pos)
}

// SetFlag sets the given flag.
func (bitmask BitMask) SetFlag(flag BitMask) BitMask {
	return bitmask | flag
}

// ClearFlag clears the bit at the given position.
func (bitmask BitMask) ClearBit(pos uint) BitMask {
	return bitmask.ClearFlag(1 << pos)
}

// ClearFlag clears the given flag.
func (bitmask BitMask) ClearFlag(flag BitMask) BitMask {
	return bitmask & ^flag
}

// HasFlag checks whether the flag at the given position is set.
func (bitmask BitMask) HasFlag(pos uint) bool {
	return bitmask&(1<<pos) > 0
}

// ModifyFlag sets or clears the flag at the given position, given the supplied state bool.
func (bitmask BitMask) ModifyFlag(pos uint, state bool) BitMask {
	if state {
		return bitmask.SetBit(pos)
	}
	return bitmask.ClearBit(pos)
}
