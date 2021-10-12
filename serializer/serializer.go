package serializer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

type (
	// ArrayOf32Bytes is an array of 32 bytes.
	ArrayOf32Bytes = [32]byte

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

	// WrittenObjectConsumer gets called after an object has been serialized into a Serializer.
	WrittenObjectConsumer func(index int, written []byte) error

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
func (s *Serializer) Write32BytesArraySlice(data SliceOfArraysOf32Bytes, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, arrayRules *ArrayRules, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	sliceLength := len(data)

	var arrayElementValidator ElementValidationFunc
	if arrayRules != nil && deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := arrayRules.CheckBounds(uint(sliceLength)); err != nil {
			s.err = errProducer(err)
			return s
		}

		arrayElementValidator = arrayRules.ElementValidationFunc(arrayRules.ValidationMode)
	}

	_ = s.writeSliceLength(sliceLength, lenType, errProducer)
	if s.err != nil {
		return s
	}

	for i := range data {
		element := data[i][:]

		if arrayElementValidator != nil {
			if err := arrayElementValidator(i, element); err != nil {
				s.err = errProducer(err)
				return s
			}
		}

		if _, err := s.buf.Write(element); err != nil {
			s.err = errProducer(err)
			return s
		}
	}
	return s
}

// Write64BytesArraySlice writes a slice of arrays of 64 bytes to the Serializer.
func (s *Serializer) Write64BytesArraySlice(data SliceOfArraysOf64Bytes, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, arrayRules *ArrayRules, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	sliceLength := len(data)

	var arrayElementValidator ElementValidationFunc
	if arrayRules != nil && deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := arrayRules.CheckBounds(uint(sliceLength)); err != nil {
			s.err = errProducer(err)
			return s
		}

		arrayElementValidator = arrayRules.ElementValidationFunc(arrayRules.ValidationMode)
	}

	_ = s.writeSliceLength(sliceLength, lenType, errProducer)
	if s.err != nil {
		return s
	}

	for i := range data {
		element := data[i][:]

		if arrayElementValidator != nil {
			if err := arrayElementValidator(i, element); err != nil {
				s.err = errProducer(err)
				return s
			}
		}

		if _, err := s.buf.Write(element); err != nil {
			s.err = errProducer(err)
			return s
		}
	}
	return s
}

// WriteObject writes the given Serializable to the Serializer.
func (s *Serializer) WriteObject(seri Serializable, deSeriMode DeSerializationMode, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
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

// WriteSliceOfObjects writes Serializables into the Serializer.
// For every written Serializable, the given WrittenObjectConsumer is called if it isn't nil.
func (s *Serializer) WriteSliceOfObjects(seris Serializables, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, woc WrittenObjectConsumer, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	_ = s.writeSliceLength(len(seris), lenType, errProducer)
	if s.err != nil {
		return s
	}

	for i, seri := range seris {
		ser, err := seri.Serialize(deSeriMode)
		if err != nil {
			s.err = errProducer(err)
			return s
		}
		if _, err := s.buf.Write(ser); err != nil {
			s.err = errProducer(err)
			return s
		}
		if woc != nil {
			if err := woc(i, ser); err != nil {
				s.err = errProducer(err)
				return s
			}
		}
	}
	return s
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
func (s *Serializer) WritePayload(payload Serializable, deSeriMode DeSerializationMode, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	if payload == nil {
		if err := binary.Write(&s.buf, binary.LittleEndian, uint32(0)); err != nil {
			s.err = errProducer(fmt.Errorf("unable to serialize zero payload length: %w", err))
		}
		return s
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
// Use this function only to read fixed size slices/arrays, otherwise
// use ReadVariableByteSlice instead.
func (d *Deserializer) ReadBytes(slice *[]byte, numBytes int, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	dest := make([]byte, numBytes)
	if len(d.src) < numBytes {
		d.err = errProducer(ErrDeserializationNotEnoughData)
		return d
	}

	copy(dest, d.src[:numBytes])
	*slice = dest

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

		arrayElementValidator = arrayRules.ElementValidationFunc(arrayRules.ValidationMode)
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

		arrayElementValidator = arrayRules.ElementValidationFunc(arrayRules.ValidationMode)
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

// ReadObject reads an object, using the given SerializableSelectorFunc.
func (d *Deserializer) ReadObject(f ReadObjectConsumerFunc, deSeriMode DeSerializationMode, typeDen TypeDenotationType, serSel SerializableSelectorFunc, errProducer ErrProducer) *Deserializer {
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

	f(seri)

	return d
}

// ReadSliceOfObjects reads a slice of objects.
func (d *Deserializer) ReadSliceOfObjects(f ReadObjectsConsumerFunc, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, typeDen TypeDenotationType, serSel SerializableSelectorFunc, arrayRules *ArrayRules, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

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

		arrayElementValidator = arrayRules.ElementValidationFunc(arrayRules.ValidationMode)
	}

	var seris Serializables
	for i := 0; i < int(sliceLength); i++ {

		// remember where we were before reading the object
		srcBefore := d.src
		offsetBefore := d.offset

		var seri Serializable
		// this mutates d.src/d.offset
		d.ReadObject(func(readSeri Serializable) { seri = readSeri }, deSeriMode, typeDen, serSel, func(err error) error {
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

	f(seris)

	return d
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
func (d *Deserializer) ReadPayload(f ReadObjectConsumerFunc, deSeriMode DeSerializationMode, sel SerializableSelectorFunc, errProducer ErrProducer) *Deserializer {
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

	f(payload)

	return d
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
