package promise

// uniqueID is a unique identifier.
type uniqueID uint64

// Next returns the next unique identifier.
func (u *uniqueID) Next() uniqueID {
	*u++

	return *u
}

// void is a function that does nothing.
func void() {}
