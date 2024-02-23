package lo

// Batch creates a new function that calls the given functions in order (nil entries are ignored).
func Batch(callbacks ...func()) func() {
	return func() {
		for _, callback := range callbacks {
			if callback != nil {
				callback()
			}
		}
	}
}

// BatchReverse creates a new function that calls the given functions in reverse order (nil entries are ignored).
func BatchReverse(callbacks ...func()) func() {
	return func() {
		for i := len(callbacks) - 1; i >= 0; i-- {
			if callback := callbacks[i]; callback != nil {
				callback()
			}
		}
	}
}
