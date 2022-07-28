package marshalutil

import (
	"fmt"
)

// MarshalUtil is a utility for reading/writing from/to a byte buffer that internally manages the offsets automatically.
type MarshalUtil struct {
	bytes       []byte
	readOffset  int
	writeOffset int
	size        int
}

// New creates a new MarshalUtil that can either be used for reading information from a slice of bytes or for writing
// information to a bytes buffer.
//
// If the MarshalUtil is supposed to read information from a slice of bytes then it receives the slice as its optional
// parameter.
//
// To create a MarshalUtil for writing information one can either create a fixed size buffer by providing the length of
// the buffer as the optional parameter or create a dynamically sized buffer by omitting the optional parameter.
func New(args ...interface{}) *MarshalUtil {
	switch argsCount := len(args); argsCount {
	case 0:
		return &MarshalUtil{
			bytes: make([]byte, 1024),
			size:  0,
		}
	case 1:
		switch param := args[0].(type) {
		case int:
			return &MarshalUtil{
				bytes: make([]byte, param),
				size:  0,
			}
		case []byte:
			return &MarshalUtil{
				bytes: param,
				size:  len(param),
			}
		default:
			panic(fmt.Errorf("illegal argument type %T in marshalutil.New(...)", param))
		}
	default:
		panic(fmt.Errorf("illegal argument count %d in marshalutil.New(...)", argsCount))
	}
}

// Write marshals the given object by writing its Bytes into the underlying buffer.
func (util *MarshalUtil) Write(object SimpleBinaryMarshaler) *MarshalUtil {
	return util.WriteBytes(object.Bytes())
}

// Parse reads information from the internal buffer by handing over the unread bytes to the passed in parser function.
func (util *MarshalUtil) Parse(parser func(data []byte) (interface{}, int, error)) (result interface{}, err error) {
	result, readBytes, err := parser(util.bytes[util.readOffset:])
	if err == nil {
		util.ReadSeek(util.readOffset + readBytes)
	}

	return
}

// ReadOffset returns the current read offset of the internal buffer.
func (util *MarshalUtil) ReadOffset() int {
	return util.readOffset
}

// WriteOffset returns the current write offset of the internal buffer.
func (util *MarshalUtil) WriteOffset() int {
	return util.writeOffset
}

// WriteSeek sets the write offset of the internal buffer. If the offset is negative then it decreases the current write
// offset instead of setting an absolute value.
func (util *MarshalUtil) WriteSeek(offset int) {
	if offset < 0 {
		util.writeOffset += offset
	} else {
		util.writeOffset = offset
	}
}

// ReadSeek sets the read offset of the internal buffer. If the offset is negative then it decreases the current read
// offset instead of setting an absolute value.
func (util *MarshalUtil) ReadSeek(offset int) {
	if offset < 0 {
		util.readOffset += offset
	} else {
		util.readOffset = offset
	}
}

// Bytes returns the internal buffer. If the optional clone parameter is set to true, then the buffer is cloned before
// being returned.
func (util *MarshalUtil) Bytes(clone ...bool) []byte {
	if len(clone) >= 1 && clone[0] {
		clone := make([]byte, util.size)
		copy(clone, util.bytes)

		return clone
	}

	return util.bytes[:util.size]
}

// DoneReading checks if there are any bytes left to read.
func (util *MarshalUtil) DoneReading() (bool, error) {
	_, err := util.checkReadCapacity(0)
	return util.ReadOffset() == util.size, err
}

// checkReadCapacity checks if the internal buffer has enough bytes left to successfully read the given length.
func (util *MarshalUtil) checkReadCapacity(length int) (readEndOffset int, err error) {
	readEndOffset = util.readOffset + length

	if readEndOffset > util.size {
		err = fmt.Errorf("tried to read %d bytes from %d bytes input", readEndOffset, util.size)
	}

	return
}

// expandWriteCapacity expands the internal buffer to have enough space to write the given amount of bytes.
func (util *MarshalUtil) expandWriteCapacity(length int) (writeEndOffset int) {
	writeEndOffset = util.writeOffset + length

	if writeEndOffset > len(util.bytes) {
		extendedBytes := make([]byte, writeEndOffset-len(util.bytes))
		util.bytes = append(util.bytes, extendedBytes...)
	}
	util.size = writeEndOffset

	return
}

// SimpleBinaryMarshaler represents objects that have a Bytes method for marshaling. In contrast to go's built marshaler
// interface (encoding.BinaryMarshaler) this interface expect no errors to be returned.
type SimpleBinaryMarshaler interface {
	// Bytes returns a marshaled version of the object.
	Bytes() []byte
}
