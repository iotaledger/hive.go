package byteutils

func ReadAvailableBytesToBuffer(target []byte, targetOffset int, source []byte, sourceOffset int, sourceLength int) int {
	availableBytes := sourceLength - sourceOffset
	requiredBytes := len(target) - targetOffset

	var bytesToRead int
	if availableBytes < requiredBytes {
		bytesToRead = availableBytes
	} else {
		bytesToRead = requiredBytes
	}

	copy(target[targetOffset:], source[sourceOffset:sourceOffset+bytesToRead])

	return bytesToRead
}

// Concat concatenates the byte slices into a new byte slice.
func Concat(byteSlices ...[]byte) (result []byte) {
	// sanitize parameters
	if len(byteSlices) == 0 {
		panic("calls to Concat require at least one argument")
	}

	// concat byte slices
	for _, byteSlice := range byteSlices {
		result = append(result, byteSlice...)
	}

	return
}
