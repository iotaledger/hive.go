package serializer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sort"
	"time"
)

type (
	// ArrayOf12Bytes is an array of 12 bytes.
	ArrayOf12Bytes = [12]byte

	// ArrayOf20Bytes is an array of 20 bytes.
	ArrayOf20Bytes = [20]byte

	// ArrayOf32Bytes is an array of 32 bytes.
	ArrayOf32Bytes = [32]byte

	// ArrayOf38Bytes is an array of 38 bytes.
	ArrayOf38Bytes = [38]byte

	// ArrayOf64Bytes is an array of 64 bytes.
	ArrayOf64Bytes = [64]byte

	// ArrayOf49Bytes is an array of 49 bytes.
	ArrayOf49Bytes = [49]byte

	// SliceOfArraysOf32Bytes is a slice of arrays of which each is 32 bytes.
	SliceOfArraysOf32Bytes = []ArrayOf32Bytes

	// SliceOfArraysOf64Bytes is a slice of arrays of which each is 64 bytes.
	SliceOfArraysOf64Bytes = []ArrayOf64Bytes

	// ErrProducer produces an error.
	ErrProducer func(err error) error

	// ErrProducerWithLeftOver produces an error and is called with the bytes left to read.
	ErrProducerWithLeftOver func(left int, err error) error

	// ReadObjectConsumerFunc gets called after an object has been deserialized from a Deserializer.
	ReadObjectConsumerFunc func(seri Serializable)

	// ReadObjectsConsumerFunc gets called after objects have been deserialized from a Deserializer.
	ReadObjectsConsumerFunc func(seri Serializables)
)

// SeriLengthPrefixType defines the type of the value denoting the length of a collection.
type SeriLengthPrefixType byte

const (
	// SeriLengthPrefixTypeAsByte defines a collection length to be denoted by a byte.
	SeriLengthPrefixTypeAsByte SeriLengthPrefixType = iota
	// SeriLengthPrefixTypeAsUint16 defines a collection length to be denoted by a uint16.
	SeriLengthPrefixTypeAsUint16
	// SeriLengthPrefixTypeAsUint32 defines a collection length to be denoted by a uint32.
	SeriLengthPrefixTypeAsUint32
)

// NewSerializer creates a new Serializer.
func NewSerializer() *Serializer {
	return &Serializer{}
}

// Serializer is a utility to serialize bytes.
type Serializer struct {
	buf bytes.Buffer
	err error
}

// Serialize finishes the serialization by returning the serialized bytes
// or an error if any intermediate step created one.
func (s *Serializer) Serialize() ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.buf.Bytes(), nil
}

// AbortIf calls the given ErrProducer if the Serializer did not encounter an error yet.
// Return nil from the ErrProducer to indicate continuation of the serialization.
func (s *Serializer) AbortIf(errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}
	if err := errProducer(nil); err != nil {
		s.err = err
	}
	return s
}

// Do calls f in the Serializer chain.
func (s *Serializer) Do(f func()) *Serializer {
	if s.err != nil {
		return s
	}
	f()
	return s
}

// Written returns the amount of bytes written into the Serializer.
func (s *Serializer) Written() int {
	return s.buf.Len()
}

// WriteNum writes the given num v to the Serializer.
func (s *Serializer) WriteNum(v interface{}, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}
	if err := binary.Write(&s.buf, binary.LittleEndian, v); err != nil {
		s.err = errProducer(err)
	}
	return s
}

// WriteUint256 writes the given *big.Int v representing an uint256 value to the Serializer.
func (s *Serializer) WriteUint256(num *big.Int, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	if num == nil {
		s.err = errProducer(ErrUint256Nil)
		return s
	}

	switch {
	case num.Sign() == -1:
		s.err = errProducer(ErrUint256NumNegative)
		return s
	case len(num.Bytes()) > UInt256ByteSize:
		s.err = errProducer(ErrUint256TooBig)
		return s
	}

	numBytes := num.Bytes()

	// order to little endianness
	for i, j := 0, len(numBytes)-1; i < j; i, j = i+1, j-1 {
		numBytes[i], numBytes[j] = numBytes[j], numBytes[i]
	}

	padded := append(numBytes, make([]byte, 32-len(numBytes))...)

	if _, err := s.buf.Write(padded); err != nil {
		s.err = errProducer(err)
		return s
	}

	return s
}

// WriteBool writes the given bool to the Serializer.
func (s *Serializer) WriteBool(v bool, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	var val byte
	if v {
		val = 1
	}

	if err := s.buf.WriteByte(val); err != nil {
		s.err = errProducer(err)
	}
	return s
}

// WriteByte writes the given byte to the Serializer.
func (s *Serializer) WriteByte(data byte, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}
	if err := s.buf.WriteByte(data); err != nil {
		s.err = errProducer(err)
	}
	return s
}

// WriteBytes writes the given byte slice to the Serializer.
// Use this function only to write fixed size slices/arrays, otherwise
// use WriteVariableByteSlice instead.
func (s *Serializer) WriteBytes(data []byte, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}
	if _, err := s.buf.Write(data); err != nil {
		s.err = errProducer(err)
	}
	return s
}

// writes the given length to the Serializer as the defined SeriLengthPrefixType.
func (s *Serializer) writeSliceLength(l int, lenType SeriLengthPrefixType, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}
	switch lenType {
	case SeriLengthPrefixTypeAsByte:
		if l > math.MaxUint8 {
			s.err = errProducer(fmt.Errorf("unable to serialize collection length: length %d is out of range (0-%d)", l, math.MaxUint8))
			return s
		}
		if err := s.buf.WriteByte(byte(l)); err != nil {
			s.err = errProducer(err)
			return s
		}
	case SeriLengthPrefixTypeAsUint16:
		if l > math.MaxUint16 {
			s.err = errProducer(fmt.Errorf("unable to serialize collection length: length %d is out of range (0-%d)", l, math.MaxUint16))
			return s
		}
		if err := binary.Write(&s.buf, binary.LittleEndian, uint16(l)); err != nil {
			s.err = errProducer(err)
			return s
		}
	case SeriLengthPrefixTypeAsUint32:
		if l > math.MaxUint32 {
			s.err = errProducer(fmt.Errorf("unable to serialize collection length: length %d is out of range (0-%d)", l, math.MaxUint32))
			return s
		}
		if err := binary.Write(&s.buf, binary.LittleEndian, uint32(l)); err != nil {
			s.err = errProducer(err)
			return s
		}
	default:
		panic(fmt.Sprintf("unknown slice length type %v", lenType))
	}
	return s
}

// WriteVariableByteSlice writes the given slice with its length to the Serializer.
func (s *Serializer) WriteVariableByteSlice(data []byte, lenType SeriLengthPrefixType, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	_ = s.writeSliceLength(len(data), lenType, errProducer)
	if s.err != nil {
		return s
	}

	if _, err := s.buf.Write(data); err != nil {
		s.err = errProducer(err)
		return s
	}
	return s
}

// Write32BytesArraySlice writes a slice of arrays of 32 bytes to the Serializer.
func (s *Serializer) Write32BytesArraySlice(slice SliceOfArraysOf32Bytes, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, arrayRules *ArrayRules, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	data := make([][]byte, len(slice))
	for i := range slice {
		data[i] = slice[i][:]
	}

	return s.writeSliceOfByteSlices(data, deSeriMode, lenType, arrayRules, errProducer)
}

// Write64BytesArraySlice writes a slice of arrays of 64 bytes to the Serializer.
func (s *Serializer) Write64BytesArraySlice(slice SliceOfArraysOf64Bytes, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, arrayRules *ArrayRules, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	data := make([][]byte, len(slice))
	for i := range slice {
		data[i] = slice[i][:]
	}

	return s.writeSliceOfByteSlices(data, deSeriMode, lenType, arrayRules, errProducer)
}

// WriteSliceOfObjects writes Serializables into the Serializer.
// For every written Serializable, the given WrittenObjectConsumer is called if it isn't nil.
func (s *Serializer) WriteSliceOfObjects(source interface{}, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, arrayRules *ArrayRules, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	seris := s.sourceToSerializables(source)

	data := make([][]byte, len(seris))
	for i, seri := range seris {
		if deSeriMode.HasMode(DeSeriModePerformValidation) && arrayRules.Guards.WriteGuard != nil {
			if err := arrayRules.Guards.WriteGuard(seri); err != nil {
				s.err = errProducer(err)
				return s
			}
		}
		ser, err := seri.Serialize(deSeriMode)
		if err != nil {
			s.err = errProducer(err)
			return s
		}
		data[i] = ser
	}

	return s.writeSliceOfByteSlices(data, deSeriMode, lenType, arrayRules, errProducer)
}

func (s *Serializer) writeSliceOfByteSlices(data [][]byte, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, sliceRules *ArrayRules, errProducer ErrProducer) *Serializer {
	var eleValFunc ElementValidationFunc
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := sliceRules.CheckBounds(uint(len(data))); err != nil {
			s.err = errProducer(err)
			return s
		}
		eleValFunc = sliceRules.ElementValidationFunc()
	}

	_ = s.writeSliceLength(len(data), lenType, errProducer)
	if s.err != nil {
		return s
	}

	// we only auto sort if the rules require it
	if deSeriMode.HasMode(DeSeriModePerformLexicalOrdering) && sliceRules.ValidationMode.HasMode(ArrayValidationModeLexicalOrdering) {
		sort.Slice(data, func(i, j int) bool {
			return bytes.Compare(data[i], data[j]) < 0
		})
	}

	for i, ele := range data {
		if eleValFunc != nil {
			if err := eleValFunc(i, ele); err != nil {
				s.err = errProducer(err)
				return s
			}
		}
		if _, err := s.buf.Write(ele); err != nil {
			s.err = errProducer(err)
			return s
		}
	}

	return s
}

// WriteObject writes the given Serializable to the Serializer.
func (s *Serializer) WriteObject(seri Serializable, deSeriMode DeSerializationMode, guard SerializableWriteGuardFunc, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := guard(seri); err != nil {
			s.err = errProducer(err)
			return s
		}
	}

	seriBytes, err := seri.Serialize(deSeriMode)
	if err != nil {
		s.err = errProducer(err)
		return s
	}

	if _, err := s.buf.Write(seriBytes); err != nil {
		s.err = errProducer(err)
	}

	return s
}

func (s *Serializer) sourceToSerializables(source interface{}) Serializables {
	var seris Serializables
	switch x := source.(type) {
	case Serializables:
		seris = x
	case SerializableSlice:
		seris = x.ToSerializables()
	default:
		panic(fmt.Sprintf("invalid source: %T", source))
	}
	return seris
}

// WriteTime writes a marshaled Time value to the internal buffer.
func (s *Serializer) WriteTime(timeToWrite time.Time, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	nanoSeconds := timeToWrite.UnixNano()
	timeToWrite.IsZero()
	// the zero value of time translates to -6795364578871345152
	if nanoSeconds == -6795364578871345152 {
		if err := binary.Write(&s.buf, binary.LittleEndian, nanoSeconds); err != nil {
			s.err = errProducer(err)
		}
	} else {
		if err := binary.Write(&s.buf, binary.LittleEndian, nanoSeconds); err != nil {
			s.err = errProducer(err)
		}
	}
	return s
}

// WritePayload writes the given payload Serializable into the Serializer.
// This is different to WriteObject as it also writes the length denotation of the payload.
func (s *Serializer) WritePayload(payload Serializable, deSeriMode DeSerializationMode, guard SerializableWriteGuardFunc, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	if payload == nil {
		if err := binary.Write(&s.buf, binary.LittleEndian, uint32(0)); err != nil {
			s.err = errProducer(fmt.Errorf("unable to serialize zero payload length: %w", err))
		}
		return s
	}

	if guard != nil {
		if err := guard(payload); err != nil {
			s.err = errProducer(err)
			return s
		}
	}

	payloadBytes, err := payload.Serialize(deSeriMode)
	if err != nil {
		s.err = errProducer(fmt.Errorf("unable to serialize payload: %w", err))
		return s
	}

	if err := binary.Write(&s.buf, binary.LittleEndian, uint32(len(payloadBytes))); err != nil {
		s.err = errProducer(fmt.Errorf("unable to serialize payload length: %w", err))
		return s
	}

	if _, err := s.buf.Write(payloadBytes); err != nil {
		s.err = errProducer(err)
	}

	return s
}

// WriteString writes the given string to the Serializer.
func (s *Serializer) WriteString(str string, lenType SeriLengthPrefixType, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	_ = s.writeSliceLength(len(str), lenType, errProducer)
	if s.err != nil {
		return s
	}

	if _, err := s.buf.Write([]byte(str)); err != nil {
		s.err = errProducer(err)
	}

	return s
}

// NewDeserializer creates a new Deserializer.
func NewDeserializer(src []byte) *Deserializer {
	return &Deserializer{src: src}
}

// Deserializer is a utility to deserialize bytes.
type Deserializer struct {
	src    []byte
	offset int
	err    error
}

// Skip skips the number of bytes during deserialization.
func (d *Deserializer) Skip(skip int, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	if len(d.src) < skip {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}
	d.offset += skip
	d.src = d.src[skip:]
	return d
}

// ReadBool reads a bool into dest.
func (d *Deserializer) ReadBool(dest *bool, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	if len(d.src) == 0 {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	switch d.src[0] {
	case 0:
		*dest = false
	case 1:
		*dest = true
	default:
		d.err = errProducer(ErrDeserializationInvalidBoolValue)
		return d
	}

	d.offset += OneByte
	d.src = d.src[OneByte:]
	return d
}

// ReadByte reads a byte into dest.
func (d *Deserializer) ReadByte(dest *byte, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	if len(d.src) == 0 {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	*dest = d.src[0]

	d.offset += OneByte
	d.src = d.src[OneByte:]
	return d
}

// ReadUint256 reads a little endian encoded uint256 into dest.
func (d *Deserializer) ReadUint256(dest *big.Int, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	if len(d.src) < UInt256ByteSize {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	source := make([]byte, UInt256ByteSize)
	copy(source, d.src[:UInt256ByteSize])

	d.offset += UInt256ByteSize
	d.src = d.src[UInt256ByteSize:]

	// convert to big endian
	for i, j := 0, len(source)-1; i < j; i, j = i+1, j-1 {
		source[i], source[j] = source[j], source[i]
	}

	dest.SetBytes(source)

	return d
}

// ReadNum reads a number into dest.
func (d *Deserializer) ReadNum(dest interface{}, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	l := len(d.src)

	switch x := dest.(type) {
	case *uint8:
		if l < OneByte {
			d.err = errProducer(ErrDeserializationNotEnoughData)
			return d
		}
		l = OneByte
		*x = d.src[0]

	case *uint16:
		if l < UInt16ByteSize {
			d.err = errProducer(ErrDeserializationNotEnoughData)
			return d
		}
		l = UInt16ByteSize
		*x = binary.LittleEndian.Uint16(d.src[:UInt16ByteSize])

	case *uint32:
		if l < UInt32ByteSize {
			d.err = errProducer(ErrDeserializationNotEnoughData)
			return d
		}
		l = UInt32ByteSize
		*x = binary.LittleEndian.Uint32(d.src[:UInt32ByteSize])
	case *uint64:
		if l < UInt64ByteSize {
			d.err = errProducer(ErrDeserializationNotEnoughData)
			return d
		}
		l = UInt64ByteSize
		*x = binary.LittleEndian.Uint64(d.src[:UInt64ByteSize])

	default:
		panic(fmt.Sprintf("unsupported ReadNum type %T", dest))
	}

	d.offset += l
	d.src = d.src[l:]

	return d
}

// ReadBytes reads specified number of bytes.
// Use this function only to read fixed size slices/arrays, otherwise use ReadVariableByteSlice instead.
func (d *Deserializer) ReadBytes(slice *[]byte, numBytes int, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	if len(d.src) < numBytes {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	dest := make([]byte, numBytes)

	copy(dest, d.src[:numBytes])
	*slice = dest

	d.offset += numBytes
	d.src = d.src[numBytes:]

	return d
}

// ReadBytesInPlace reads slice length amount of bytes into slice.
// Use this function only to read arrays.
func (d *Deserializer) ReadBytesInPlace(slice []byte, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	numBytes := len(slice)
	if len(d.src) < numBytes {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	copy(slice, d.src[:numBytes])

	d.offset += numBytes
	d.src = d.src[numBytes:]

	return d
}

// ReadVariableByteSlice reads a variable byte slice which is denoted by the given SeriLengthPrefixType.
func (d *Deserializer) ReadVariableByteSlice(slice *[]byte, lenType SeriLengthPrefixType, errProducer ErrProducer, maxRead ...int) *Deserializer {
	if d.err != nil {
		return d
	}

	sliceLength, err := d.readSliceLength(lenType, errProducer)
	if err != nil {
		d.err = err
		return d
	}

	if len(maxRead) > 0 && sliceLength > maxRead[0] {
		d.err = errProducer(fmt.Errorf("%w: denoted %d bytes, max allowed %d ", ErrDeserializationLengthInvalid, sliceLength, maxRead[0]))
		return d
	}
	dest := make([]byte, sliceLength)
	if len(d.src) < sliceLength {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	copy(dest, d.src[:sliceLength])
	*slice = dest

	d.offset += sliceLength
	d.src = d.src[sliceLength:]

	return d
}

// ReadArrayOf12Bytes reads an array of 12 bytes.
func (d *Deserializer) ReadArrayOf12Bytes(arr *ArrayOf12Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 12

	l := len(d.src)
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	copy(arr[:], d.src[:length])
	d.offset += length
	d.src = d.src[length:]

	return d
}

// ReadArrayOf20Bytes reads an array of 20 bytes.
func (d *Deserializer) ReadArrayOf20Bytes(arr *ArrayOf20Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 20

	l := len(d.src)
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	copy(arr[:], d.src[:length])
	d.offset += length
	d.src = d.src[length:]

	return d
}

// ReadArrayOf32Bytes reads an array of 32 bytes.
func (d *Deserializer) ReadArrayOf32Bytes(arr *ArrayOf32Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 32

	l := len(d.src)
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	copy(arr[:], d.src[:length])
	d.offset += length
	d.src = d.src[length:]

	return d
}

// ReadArrayOf38Bytes reads an array of 38 bytes.
func (d *Deserializer) ReadArrayOf38Bytes(arr *ArrayOf38Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 38

	l := len(d.src)
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	copy(arr[:], d.src[:length])
	d.offset += length
	d.src = d.src[length:]

	return d
}

// ReadArrayOf64Bytes reads an array of 64 bytes.
func (d *Deserializer) ReadArrayOf64Bytes(arr *ArrayOf64Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 64

	l := len(d.src)
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	copy(arr[:], d.src[:length])
	d.offset += length
	d.src = d.src[length:]

	return d
}

// ReadArrayOf49Bytes reads an array of 49 bytes.
func (d *Deserializer) ReadArrayOf49Bytes(arr *ArrayOf49Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 49

	l := len(d.src)
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	copy(arr[:], d.src[:length])
	d.offset += length
	d.src = d.src[length:]

	return d
}

// reads the length of a slice.
func (d *Deserializer) readSliceLength(lenType SeriLengthPrefixType, errProducer ErrProducer) (int, error) {
	l := len(d.src)
	var sliceLength int

	switch lenType {

	case SeriLengthPrefixTypeAsByte:
		if l < OneByte {
			return 0, errProducer(ErrDeserializationNotEnoughData)
		}
		l = OneByte
		sliceLength = int(d.src[0])

	case SeriLengthPrefixTypeAsUint16:
		if l < UInt16ByteSize {
			return 0, errProducer(ErrDeserializationNotEnoughData)
		}
		l = UInt16ByteSize
		sliceLength = int(binary.LittleEndian.Uint16(d.src[:UInt16ByteSize]))

	case SeriLengthPrefixTypeAsUint32:
		if l < UInt32ByteSize {
			return 0, errProducer(ErrDeserializationNotEnoughData)
		}
		l = UInt32ByteSize
		sliceLength = int(binary.LittleEndian.Uint32(d.src[:UInt32ByteSize]))

	default:
		panic(fmt.Sprintf("unknown slice length type %v", lenType))
	}

	d.offset += l
	d.src = d.src[l:]

	return sliceLength, nil
}

// ReadSliceOfArraysOf32Bytes reads a slice of arrays of 32 bytes.
func (d *Deserializer) ReadSliceOfArraysOf32Bytes(slice *SliceOfArraysOf32Bytes, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, arrayRules *ArrayRules, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 32

	sliceLength, err := d.readSliceLength(lenType, errProducer)
	if err != nil {
		d.err = err
		return d
	}

	var arrayElementValidator ElementValidationFunc
	if arrayRules != nil && deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := arrayRules.CheckBounds(uint(sliceLength)); err != nil {
			d.err = errProducer(err)
			return d
		}

		arrayElementValidator = arrayRules.ElementValidationFunc()
	}

	s := make(SliceOfArraysOf32Bytes, sliceLength)
	for i := 0; i < sliceLength; i++ {
		if len(d.src) < length {
			d.err = errProducer(ErrDeserializationNotEnoughData)
			return d
		}

		if arrayElementValidator != nil {
			if err := arrayElementValidator(i, d.src[:length]); err != nil {
				d.err = errProducer(err)
				return d
			}
		}

		copy(s[i][:], d.src[:length])
		d.offset += length
		d.src = d.src[length:]
	}

	*slice = s

	return d
}

// ReadSliceOfArraysOf64Bytes reads a slice of arrays of 64 bytes.
func (d *Deserializer) ReadSliceOfArraysOf64Bytes(slice *SliceOfArraysOf64Bytes, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, arrayRules *ArrayRules, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 64

	sliceLength, err := d.readSliceLength(lenType, errProducer)
	if err != nil {
		d.err = err
		return d
	}

	var arrayElementValidator ElementValidationFunc
	if arrayRules != nil && deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := arrayRules.CheckBounds(uint(sliceLength)); err != nil {
			d.err = errProducer(err)
			return d
		}

		arrayElementValidator = arrayRules.ElementValidationFunc()
	}

	s := make(SliceOfArraysOf64Bytes, sliceLength)
	for i := 0; i < sliceLength; i++ {
		if len(d.src) < length {
			d.err = errProducer(ErrDeserializationNotEnoughData)
			return d
		}

		if arrayElementValidator != nil {
			if err := arrayElementValidator(i, d.src[:length]); err != nil {
				d.err = errProducer(err)
				return d
			}
		}

		copy(s[i][:], d.src[:length])
		d.offset += length
		d.src = d.src[length:]
	}

	*slice = s

	return d
}

// ReadObject reads an object, using the given SerializableReadGuardFunc.
func (d *Deserializer) ReadObject(target interface{}, deSeriMode DeSerializationMode, typeDen TypeDenotationType, serSel SerializableReadGuardFunc, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	l := len(d.src)
	var ty uint32
	switch typeDen {
	case TypeDenotationUint32:
		if l < UInt32ByteSize {
			d.err = errProducer(ErrDeserializationNotEnoughData)
			return d
		}
		ty = binary.LittleEndian.Uint32(d.src)
	case TypeDenotationByte:
		if l < OneByte {
			d.err = errProducer(ErrDeserializationNotEnoughData)
			return d
		}
		ty = uint32(d.src[0])
	case TypeDenotationNone:
		// object has no type denotation
	}

	seri, err := serSel(ty)
	if err != nil {
		d.err = errProducer(err)
		return d
	}

	bytesConsumed, err := seri.Deserialize(d.src, deSeriMode)
	if err != nil {
		d.err = errProducer(err)
		return d
	}

	d.offset += bytesConsumed
	d.src = d.src[bytesConsumed:]

	d.readSerializableIntoTarget(target, seri)

	return d
}

// ReadSliceOfObjects reads a slice of objects.
func (d *Deserializer) ReadSliceOfObjects(target interface{}, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, typeDen TypeDenotationType, arrayRules *ArrayRules, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	sliceLength, err := d.readSliceLength(lenType, errProducer)
	if err != nil {
		d.err = err
		return d
	}

	var arrayElementValidator ElementValidationFunc
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := arrayRules.CheckBounds(uint(sliceLength)); err != nil {
			d.err = errProducer(err)
			return d
		}

		arrayElementValidator = arrayRules.ElementValidationFunc()
	}

	var seris Serializables
	for i := 0; i < int(sliceLength); i++ {

		// remember where we were before reading the object
		srcBefore := d.src
		offsetBefore := d.offset

		var seri Serializable
		// this mutates d.src/d.offset
		d.ReadObject(func(readSeri Serializable) { seri = readSeri }, deSeriMode, typeDen, arrayRules.Guards.ReadGuard, func(err error) error {
			return errProducer(err)
		})

		// there was an error
		if seri == nil {
			return d
		}

		bytesConsumed := d.offset - offsetBefore

		if arrayElementValidator != nil {
			if err := arrayElementValidator(i, srcBefore[:bytesConsumed]); err != nil {
				d.err = errProducer(err)
				return d
			}
		}

		seris = append(seris, seri)
	}

	d.readSerializablesIntoTarget(target, seris)

	return d
}

func (d *Deserializer) readSerializablesIntoTarget(target interface{}, seris Serializables) {
	switch x := target.(type) {
	case func(seri Serializables):
		x(seris)
	case SerializableSlice:
		x.FromSerializables(seris)
	default:
		panic("invalid target")
	}
}

// ReadTime reads a Time value from the internal buffer.
func (d *Deserializer) ReadTime(dest *time.Time, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	l := len(d.src)

	if l < Int64ByteSize {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}
	l = Int64ByteSize
	nanoSeconds := int64(binary.LittleEndian.Uint64(d.src[:Int64ByteSize]))
	if nanoSeconds == 0 {
		*dest = time.Time{}
	} else {
		*dest = time.Unix(0, nanoSeconds)
	}

	d.offset += l
	d.src = d.src[l:]

	return d
}

// ReadPayload reads a payload.
func (d *Deserializer) ReadPayload(s interface{}, deSeriMode DeSerializationMode, sel SerializableReadGuardFunc, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	if len(d.src) < PayloadLengthByteSize {
		d.err = errProducer(fmt.Errorf("%w: data is smaller than payload length denotation", ErrDeserializationNotEnoughData))
		return d
	}

	payloadLength := binary.LittleEndian.Uint32(d.src)
	d.offset += PayloadLengthByteSize
	d.src = d.src[PayloadLengthByteSize:]

	// nothing to do
	if payloadLength == 0 {
		return d
	}

	switch {
	case len(d.src) < MinPayloadByteSize:
		d.err = errProducer(fmt.Errorf("%w: payload data is smaller than min. required length %d", ErrDeserializationNotEnoughData, MinPayloadByteSize))
		return d
	case len(d.src) < int(payloadLength):
		d.err = errProducer(fmt.Errorf("%w: payload length denotes more bytes than are available", ErrDeserializationNotEnoughData))
		return d
	}

	payload, err := sel(binary.LittleEndian.Uint32(d.src))
	if err != nil {
		d.err = errProducer(err)
		return d
	}

	payloadBytesConsumed, err := payload.Deserialize(d.src, deSeriMode)
	if err != nil {
		d.err = errProducer(err)
		return d
	}

	if payloadBytesConsumed != int(payloadLength) {
		d.err = errProducer(fmt.Errorf("%w: denoted payload length (%d) doesn't equal the size of deserialized payload (%d)", ErrInvalidBytes, payloadLength, payloadBytesConsumed))
		return d
	}

	d.offset += payloadBytesConsumed
	d.src = d.src[payloadBytesConsumed:]

	d.readSerializableIntoTarget(s, payload)

	return d
}

func (d *Deserializer) readSerializableIntoTarget(target interface{}, s Serializable) {
	switch t := target.(type) {
	case *Serializable:
		*t = s
	case func(seri Serializable):
		t(s)
	default:
		if reflect.TypeOf(target).Kind() != reflect.Ptr {
			panic("target parameter must be pointer or Serializable")
		}
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(s))
	}
}

// ReadString reads a string.
func (d *Deserializer) ReadString(s *string, lenType SeriLengthPrefixType, errProducer ErrProducer, maxSize ...int) *Deserializer {
	if d.err != nil {
		return d
	}

	strLen, err := d.readSliceLength(lenType, errProducer)
	if err != nil {
		d.err = err
		return d
	}

	if len(maxSize) > 0 && strLen > maxSize[0] {
		d.err = errProducer(fmt.Errorf("%w: string defined to be of %d bytes length but max %d is allowed", ErrDeserializationLengthInvalid, strLen, maxSize[0]))
	}

	if len(d.src) < int(strLen) {
		d.err = errProducer(fmt.Errorf("%w: data is smaller than (%d) denoted string length of %d", ErrDeserializationNotEnoughData, len(d.src), strLen))
		return d
	}

	*s = string(d.src[:strLen])

	d.offset += int(strLen)
	d.src = d.src[strLen:]

	return d
}

// AbortIf calls the given ErrProducer if the Deserializer did not encounter an error yet.
// Return nil from the ErrProducer to indicate continuation of the deserialization.
func (d *Deserializer) AbortIf(errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	if err := errProducer(nil); err != nil {
		d.err = err
	}
	return d
}

// CheckTypePrefix checks whether the type prefix corresponds to the expected given prefix.
// This function will advance the deserializer by the given TypeDenotationType length.
func (d *Deserializer) CheckTypePrefix(prefix uint32, prefixType TypeDenotationType, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	var toSkip int
	switch prefixType {
	case TypeDenotationUint32:
		if err := CheckType(d.src, prefix); err != nil {
			d.err = errProducer(err)
			return d
		}
		toSkip = UInt32ByteSize
	case TypeDenotationByte:
		if err := CheckTypeByte(d.src, byte(prefix)); err != nil {
			d.err = errProducer(err)
			return d
		}
		toSkip = OneByte
	default:
		panic("invalid type prefix in CheckTypePrefix()")
	}

	return d.Skip(toSkip, func(err error) error { return err })
}

// Do calls f in the Deserializer chain.
func (d *Deserializer) Do(f func()) *Deserializer {
	if d.err != nil {
		return d
	}
	f()
	return d
}

// ConsumedAll calls the given ErrProducerWithLeftOver if not all bytes have been
// consumed from the Deserializer's src.
func (d *Deserializer) ConsumedAll(errProducer ErrProducerWithLeftOver) *Deserializer {
	if d.err != nil {
		return d
	}

	if len(d.src) != 0 {
		d.err = errProducer(len(d.src)-d.offset, ErrDeserializationNotAllConsumed)
	}

	return d
}

// Done finishes the Deserializer by returning the read bytes and occurred errors.
func (d *Deserializer) Done() (int, error) {
	return d.offset, d.err
}
