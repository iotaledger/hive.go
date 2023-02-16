package lo

func Batch(callbacks ...func()) func() {
	return func() {
		for _, callback := range callbacks {
			callback()
		}
	}
}
