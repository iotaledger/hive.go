package marshalutil

import (
	"fmt"
)

type MarshalUtil struct {
	bytes       []byte
	readOffset  int
	writeOffset int
	size        int
}

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
				size:  param,
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

func (util *MarshalUtil) Parse(parser func(data []byte) (interface{}, int, error)) (result interface{}, err error) {
	result, readBytes, err := parser(util.bytes[util.readOffset:])
	if err == nil {
		util.ReadSeek(util.readOffset + readBytes)
	}

	return
}

func (util *MarshalUtil) ReadOffset() int {
	return util.readOffset
}

func (util *MarshalUtil) WriteOffset() int {
	return util.writeOffset
}

func (util *MarshalUtil) WriteSeek(offset int) {
	if offset < 0 {
		util.writeOffset += offset
	} else {
		util.writeOffset = offset
	}
}

func (util *MarshalUtil) ReadSeek(offset int) {
	if offset < 0 {
		util.readOffset += offset
	} else {
		util.readOffset = offset
	}
}

func (util *MarshalUtil) Bytes(clone ...bool) []byte {
	if len(clone) >= 1 && clone[0] {
		clone := make([]byte, util.size)
		copy(clone, util.bytes)

		return clone
	}

	return util.bytes[:util.size]
}

func (util *MarshalUtil) checkReadCapacity(length int) (readEndOffset int, err error) {
	readEndOffset = util.readOffset + length

	if readEndOffset > util.size {
		err = fmt.Errorf("tried to read %d bytes from %d bytes input", readEndOffset, util.size)
	}

	return
}

func (util *MarshalUtil) expandWriteCapacity(length int) (writeEndOffset int) {
	writeEndOffset = util.writeOffset + length

	if writeEndOffset > util.size {
		extendedBytes := make([]byte, writeEndOffset-util.size)
		util.bytes = append(util.bytes, extendedBytes...)
		util.size = writeEndOffset
	}

	return
}

// Write marshals the given object by writing its Bytes into the underlying buffer.
func (util *MarshalUtil) Write(object SimpleBinaryMarshaler) *MarshalUtil {
	return util.WriteBytes(object.Bytes())
}

// SimpleBinaryMarshaler represents objects that have a Bytes method for marshaling. In contrast to go's built marshaler
// interface (encoding.BinaryMarshaler) this interface expect no errors to be returned.
type SimpleBinaryMarshaler interface {
	// Bytes returns a marshaled version of the object.
	Bytes() []byte
}
