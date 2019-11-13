package bitmask

type BitMask byte

// SetFlag sets the flag at the given position.
func (bitmask BitMask) SetFlag(pos uint) BitMask {
	return bitmask | (1 << pos)
}

// ClearFlag clears the flag at the given position.
func (bitmask BitMask) ClearFlag(pos uint) BitMask {
	return bitmask & ^(1 << pos)
}

// HasFlag checks whether the flag at the given position is set.
func (bitmask BitMask) HasFlag(pos uint) bool {
	return (bitmask&(1<<pos) > 0)
}

// ModifyFlag sets or clears the flag at the given position, given the supplied state bool.
func (bitmask BitMask) ModifyFlag(pos uint, state bool) BitMask {
	if state {
		return bitmask.SetFlag(pos)
	}
	return bitmask.ClearFlag(pos)
}
