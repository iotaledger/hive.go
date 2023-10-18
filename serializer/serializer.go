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

	"github.com/iotaledger/hive.go/ierrors"
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

	// ErrProducer might produce an error.
	ErrProducer func(err error) error

	// ErrProducerWithRWBytes might produce an error and is called with the currently read or written bytes.
	ErrProducerWithRWBytes func(read []byte, err error) error

	// ErrProducerWithLeftOver might produce an error and is called with the bytes left to read.
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

// WithValidation runs errProducer if deSeriMode has DeSeriModePerformValidation.
func (s *Serializer) WithValidation(deSeriMode DeSerializationMode, errProducer ErrProducerWithRWBytes) *Serializer {
	if s.err != nil {
		return s
	}
	if !deSeriMode.HasMode(DeSeriModePerformValidation) {
		return s
	}
	if err := errProducer(s.buf.Bytes(), s.err); err != nil {
		s.err = err

		return s
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

	//nolint:gocritic // false positive
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
//
//nolint:stdmethods // false positive
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
func (s *Serializer) writeSliceLength(l int, lenType SeriLengthPrefixType, errProducer ErrProducer) {
	if s.err != nil {
		return
	}
	switch lenType {
	case SeriLengthPrefixTypeAsByte:
		if l > math.MaxUint8 {
			s.err = errProducer(ierrors.Errorf("unable to serialize collection length: length %d is out of range (0-%d)", l, math.MaxUint8))

			return
		}
		if err := s.buf.WriteByte(byte(l)); err != nil {
			s.err = errProducer(err)

			return
		}
	case SeriLengthPrefixTypeAsUint16:
		if l > math.MaxUint16 {
			s.err = errProducer(ierrors.Errorf("unable to serialize collection length: length %d is out of range (0-%d)", l, math.MaxUint16))

			return
		}
		if err := binary.Write(&s.buf, binary.LittleEndian, uint16(l)); err != nil {
			s.err = errProducer(err)

			return
		}
	case SeriLengthPrefixTypeAsUint32:
		if l > math.MaxUint32 {
			s.err = errProducer(ierrors.Errorf("unable to serialize collection length: length %d is out of range (0-%d)", l, math.MaxUint32))

			return
		}
		if err := binary.Write(&s.buf, binary.LittleEndian, uint32(l)); err != nil {
			s.err = errProducer(err)

			return
		}
	default:
		panic(fmt.Sprintf("unknown slice length type %v", lenType))
	}
}

// WriteVariableByteSlice writes the given slice with its length to the Serializer.
func (s *Serializer) WriteVariableByteSlice(data []byte, lenType SeriLengthPrefixType, errProducer ErrProducer, minLen int, maxLen int) *Serializer {
	if s.err != nil {
		return s
	}

	sliceLen := len(data)
	switch {
	case maxLen > 0 && sliceLen > maxLen:
		s.err = errProducer(ierrors.Wrapf(ErrSliceLengthTooLong, "slice (len %d) exceeds max length of %d ", sliceLen, maxLen))

		return s

	case minLen > 0 && sliceLen < minLen:
		s.err = errProducer(ierrors.Wrapf(ErrSliceLengthTooShort, "slice (len %d) is less than min length of %d ", sliceLen, maxLen))

		return s
	}

	s.writeSliceLength(len(data), lenType, errProducer)
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

	return s.WriteSliceOfByteSlices(data, deSeriMode, lenType, arrayRules, errProducer)
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

	return s.WriteSliceOfByteSlices(data, deSeriMode, lenType, arrayRules, errProducer)
}

// WriteSliceOfObjects writes Serializables into the Serializer.
// For every written Serializable, the given WrittenObjectConsumer is called if it isn't nil.
func (s *Serializer) WriteSliceOfObjects(source interface{}, deSeriMode DeSerializationMode, deSeriCtx interface{}, lenType SeriLengthPrefixType, arrayRules *ArrayRules, errProducer ErrProducer) *Serializer {
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
		ser, err := seri.Serialize(deSeriMode, deSeriCtx)
		if err != nil {
			s.err = errProducer(err)

			return s
		}
		data[i] = ser
	}

	return s.WriteSliceOfByteSlices(data, deSeriMode, lenType, arrayRules, errProducer)
}

// WriteSliceOfByteSlices writes slice of []byte into the Serializer.
func (s *Serializer) WriteSliceOfByteSlices(data [][]byte, deSeriMode DeSerializationMode, lenType SeriLengthPrefixType, sliceRules *ArrayRules, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}
	var eleValFunc ElementValidationFunc
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := sliceRules.CheckBounds(uint(len(data))); err != nil {
			s.err = errProducer(err)

			return s
		}
		eleValFunc = sliceRules.ElementValidationFunc()
	}

	s.writeSliceLength(len(data), lenType, errProducer)
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
func (s *Serializer) WriteObject(seri Serializable, deSeriMode DeSerializationMode, deSeriCtx interface{}, guard SerializableWriteGuardFunc, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := guard(seri); err != nil {
			s.err = errProducer(err)

			return s
		}
	}

	seriBytes, err := seri.Serialize(deSeriMode, deSeriCtx)
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

	if err := binary.Write(&s.buf, binary.LittleEndian, TimeToUint64(timeToWrite)); err != nil {
		s.err = errProducer(err)
	}

	return s
}

// TimeToUint64 converts times to uint64 unix timestamps with second precision.
// Times before the Unix Epoch will be truncated to the Unix Epoch.
func TimeToUint64(value time.Time) uint64 {
	unixSeconds := value.Unix()

	// We don't represent times before the Unix Epoch.
	if unixSeconds < 0 {
		unixSeconds = 0
	}

	return uint64(unixSeconds)
}

// WritePayload writes the given payload Serializable into the Serializer.
// This is different to WriteObject as it also writes the length denotation of the payload.
func (s *Serializer) WritePayload(payload Serializable, deSeriMode DeSerializationMode, deSeriCtx interface{}, guard SerializableWriteGuardFunc, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}

	if payload == nil {
		if err := s.writePayloadLength(0); err != nil {
			s.err = errProducer(err)
		}

		return s
	}

	if guard != nil {
		if err := guard(payload); err != nil {
			s.err = errProducer(err)

			return s
		}
	}

	payloadBytes, err := payload.Serialize(deSeriMode, deSeriCtx)
	if err != nil {
		s.err = errProducer(ierrors.Wrap(err, "unable to serialize payload"))

		return s
	}
	if err := s.writePayloadLength(len(payloadBytes)); err != nil {
		s.err = errProducer(err)
	}

	if _, err := s.buf.Write(payloadBytes); err != nil {
		s.err = errProducer(err)
	}

	return s
}

// WritePayloadLength write payload length token into serializer.
func (s *Serializer) WritePayloadLength(length int, errProducer ErrProducer) *Serializer {
	if s.err != nil {
		return s
	}
	if err := s.writePayloadLength(length); err != nil {
		s.err = errProducer(err)
	}

	return s
}

func (s *Serializer) writePayloadLength(length int) error {
	if err := binary.Write(&s.buf, binary.LittleEndian, uint32(length)); err != nil {
		return ierrors.Wrap(err, "unable to serialize payload length")
	}

	return nil
}

// WriteString writes the given string to the Serializer.
func (s *Serializer) WriteString(str string, lenType SeriLengthPrefixType, errProducer ErrProducer, minLen int, maxLen int) *Serializer {
	if s.err != nil {
		return s
	}

	strLen := len(str)
	switch {
	case maxLen > 0 && strLen > maxLen:
		s.err = errProducer(ierrors.Wrapf(ErrStringTooLong, "string (len %d) exceeds max length of %d ", strLen, maxLen))

		return s

	case minLen > 0 && strLen < minLen:
		s.err = errProducer(ierrors.Wrapf(ErrStringTooShort, "string (len %d) is less than min length of %d", strLen, minLen))

		return s
	}

	s.writeSliceLength(strLen, lenType, errProducer)
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

func (d *Deserializer) RemainingBytes() []byte {
	return d.src[d.offset:]
}

// Skip skips the number of bytes during deserialization.
func (d *Deserializer) Skip(skip int, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	if len(d.src[d.offset:]) < skip {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}
	d.offset += skip

	return d
}

// ReadBool reads a bool into dest.
func (d *Deserializer) ReadBool(dest *bool, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	if len(d.src[d.offset:]) == 0 {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	switch d.src[d.offset : d.offset+1][0] {
	case 0:
		*dest = false
	case 1:
		*dest = true
	default:
		d.err = errProducer(ErrDeserializationInvalidBoolValue)

		return d
	}

	d.offset += OneByte

	return d
}

// ReadByte reads a byte into dest.
//
//nolint:stdmethods // false positive
func (d *Deserializer) ReadByte(dest *byte, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	if len(d.src[d.offset:]) == 0 {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	*dest = d.src[d.offset : d.offset+1][0]
	d.offset += OneByte

	return d
}

// ReadUint256 reads a little endian encoded uint256 into dest.
func (d *Deserializer) ReadUint256(dest **big.Int, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	if len(d.src[d.offset:]) < UInt256ByteSize {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	source := make([]byte, UInt256ByteSize)
	copy(source, d.src[d.offset:d.offset+UInt256ByteSize])

	d.offset += UInt256ByteSize

	// convert to big endian
	for i, j := 0, len(source)-1; i < j; i, j = i+1, j-1 {
		source[i], source[j] = source[j], source[i]
	}

	*dest = new(big.Int).SetBytes(source)

	return d
}

// numSize returns the size of the data required to represent the data when encoded.
func numSize(data any) int {
	switch data := data.(type) {
	case bool, int8, uint8, *bool, *int8, *uint8:
		return OneByte
	case int16, *int16:
		return Int16ByteSize
	case uint16, *uint16:
		return UInt16ByteSize
	case int32, *int32:
		return Int32ByteSize
	case uint32, *uint32:
		return UInt32ByteSize
	case int64, *int64:
		return Int64ByteSize
	case uint64, *uint64:
		return UInt64ByteSize
	case float32, *float32:
		return Float32ByteSize
	case float64, *float64:
		return Float64ByteSize
	default:
		panic(fmt.Sprintf("unsupported numSize type %T", data))
	}
}

// ReadNum reads a number into dest.
func (d *Deserializer) ReadNum(dest any, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	l := len(d.src[d.offset:])

	dataSize := numSize(dest)
	if l < dataSize {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}
	l = dataSize

	data := d.src[d.offset : d.offset+l]

	switch x := dest.(type) {
	case *int8:
		*x = int8(data[0])

	case *uint8:
		*x = data[0]

	case *int16:
		*x = int16(binary.LittleEndian.Uint16(data))

	case *uint16:
		*x = binary.LittleEndian.Uint16(data)

	case *int32:
		*x = int32(binary.LittleEndian.Uint32(data))

	case *uint32:
		*x = binary.LittleEndian.Uint32(data)

	case *int64:
		*x = int64(binary.LittleEndian.Uint64(data))

	case *uint64:
		*x = binary.LittleEndian.Uint64(data)

	case *float32:
		*x = math.Float32frombits(binary.LittleEndian.Uint32(data))

	case *float64:
		*x = math.Float64frombits(binary.LittleEndian.Uint64(data))

	default:
		panic(fmt.Sprintf("unsupported ReadNum type %T", dest))
	}

	d.offset += l

	return d
}

// ReadBytes reads specified number of bytes.
// Use this function only to read fixed size slices/arrays, otherwise use ReadVariableByteSlice instead.
func (d *Deserializer) ReadBytes(slice *[]byte, numBytes int, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	if len(d.src[d.offset:]) < numBytes {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	dest := make([]byte, numBytes)

	copy(dest, d.src[d.offset:d.offset+numBytes])
	*slice = dest

	d.offset += numBytes

	return d
}

// ReadBytesInPlace reads slice length amount of bytes into slice.
// Use this function only to read arrays.
func (d *Deserializer) ReadBytesInPlace(slice []byte, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}

	numBytes := len(slice)
	if len(d.src[d.offset:]) < numBytes {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	copy(slice, d.src[d.offset:d.offset+numBytes])
	d.offset += numBytes

	return d
}

// ReadVariableByteSlice reads a variable byte slice which is denoted by the given SeriLengthPrefixType.
func (d *Deserializer) ReadVariableByteSlice(slice *[]byte, lenType SeriLengthPrefixType, errProducer ErrProducer, minLen int, maxLen int) *Deserializer {
	if d.err != nil {
		return d
	}

	sliceLength, err := d.readSliceLength(lenType, errProducer)
	if err != nil {
		d.err = err

		return d
	}

	switch {
	case maxLen > 0 && sliceLength > maxLen:
		d.err = errProducer(ierrors.Wrapf(ErrDeserializationLengthInvalid, "denoted %d bytes, max allowed %d ", sliceLength, maxLen))
	case minLen > 0 && sliceLength < minLen:
		d.err = errProducer(ierrors.Wrapf(ErrDeserializationLengthInvalid, "denoted %d bytes, min required %d ", sliceLength, minLen))
	}

	dest := make([]byte, sliceLength)
	if sliceLength == 0 {
		*slice = dest

		return d
	}

	if len(d.src[d.offset:]) < sliceLength {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	copy(dest, d.src[d.offset:d.offset+sliceLength])
	*slice = dest

	d.offset += sliceLength

	return d
}

// ReadArrayOf12Bytes reads an array of 12 bytes.
func (d *Deserializer) ReadArrayOf12Bytes(arr *ArrayOf12Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 12

	l := len(d.src[d.offset:])
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	copy(arr[:], d.src[d.offset:d.offset+length])
	d.offset += length

	return d
}

// ReadArrayOf20Bytes reads an array of 20 bytes.
func (d *Deserializer) ReadArrayOf20Bytes(arr *ArrayOf20Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 20

	l := len(d.src[d.offset:])
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	copy(arr[:], d.src[d.offset:d.offset+length])
	d.offset += length

	return d
}

// ReadArrayOf32Bytes reads an array of 32 bytes.
func (d *Deserializer) ReadArrayOf32Bytes(arr *ArrayOf32Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 32

	l := len(d.src[d.offset:])
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	copy(arr[:], d.src[d.offset:d.offset+length])
	d.offset += length

	return d
}

// ReadArrayOf38Bytes reads an array of 38 bytes.
func (d *Deserializer) ReadArrayOf38Bytes(arr *ArrayOf38Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 38

	l := len(d.src[d.offset:])
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	copy(arr[:], d.src[d.offset:d.offset+length])
	d.offset += length

	return d
}

// ReadArrayOf64Bytes reads an array of 64 bytes.
func (d *Deserializer) ReadArrayOf64Bytes(arr *ArrayOf64Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 64

	l := len(d.src[d.offset:])
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	copy(arr[:], d.src[d.offset:d.offset+length])
	d.offset += length

	return d
}

// ReadArrayOf49Bytes reads an array of 49 bytes.
func (d *Deserializer) ReadArrayOf49Bytes(arr *ArrayOf49Bytes, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	const length = 49

	l := len(d.src[d.offset:])
	if l < length {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	copy(arr[:], d.src[d.offset:d.offset+length])
	d.offset += length

	return d
}

// reads the length of a slice.
func (d *Deserializer) readSliceLength(lenType SeriLengthPrefixType, errProducer ErrProducer) (int, error) {
	l := len(d.src[d.offset:])
	var sliceLength int

	switch lenType {

	case SeriLengthPrefixTypeAsByte:
		if l < OneByte {
			return 0, errProducer(ErrDeserializationNotEnoughData)
		}
		l = OneByte
		sliceLength = int(d.src[d.offset : d.offset+1][0])

	case SeriLengthPrefixTypeAsUint16:
		if l < UInt16ByteSize {
			return 0, errProducer(ErrDeserializationNotEnoughData)
		}
		l = UInt16ByteSize
		sliceLength = int(binary.LittleEndian.Uint16(d.src[d.offset : d.offset+UInt16ByteSize]))

	case SeriLengthPrefixTypeAsUint32:
		if l < UInt32ByteSize {
			return 0, errProducer(ErrDeserializationNotEnoughData)
		}
		l = UInt32ByteSize
		sliceLength = int(binary.LittleEndian.Uint32(d.src[d.offset : d.offset+UInt32ByteSize]))

	default:
		panic(fmt.Sprintf("unknown slice length type %v", lenType))
	}

	d.offset += l

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
		if len(d.src[d.offset:]) < length {
			d.err = errProducer(ErrDeserializationNotEnoughData)

			return d
		}

		if arrayElementValidator != nil {
			if err := arrayElementValidator(i, d.src[d.offset:d.offset+length]); err != nil {
				d.err = errProducer(err)

				return d
			}
		}

		copy(s[i][:], d.src[d.offset:d.offset+length])
		d.offset += length
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
		if len(d.src[d.offset:]) < length {
			d.err = errProducer(ErrDeserializationNotEnoughData)

			return d
		}

		if arrayElementValidator != nil {
			if err := arrayElementValidator(i, d.src[d.offset:d.offset+length]); err != nil {
				d.err = errProducer(err)

				return d
			}
		}

		copy(s[i][:], d.src[d.offset:d.offset+length])
		d.offset += length
	}

	*slice = s

	return d
}

// ReadObject reads an object, using the given SerializableReadGuardFunc.
func (d *Deserializer) ReadObject(target interface{}, deSeriMode DeSerializationMode, deSeriCtx interface{}, typeDen TypeDenotationType, serSel SerializableReadGuardFunc, errProducer ErrProducer) *Deserializer {
	deserializer, _ := d.readObject(target, deSeriMode, deSeriCtx, typeDen, serSel, errProducer)

	return deserializer
}

// GetObjectType reads object type but doesn't change the offset.
func (d *Deserializer) GetObjectType(typeDen TypeDenotationType) (uint32, error) {
	l := len(d.src[d.offset:])
	var ty uint32
	switch typeDen {
	case TypeDenotationUint32:
		if l < UInt32ByteSize {
			return 0, ErrDeserializationNotEnoughData
		}
		ty = binary.LittleEndian.Uint32(d.src[d.offset:])
	case TypeDenotationByte:
		if l < OneByte {
			return 0, ErrDeserializationNotEnoughData
		}
		ty = uint32(d.src[d.offset : d.offset+1][0])
	case TypeDenotationNone:
		// object has no type denotation
		return 0, nil
	}

	return ty, nil

}

func (d *Deserializer) readObject(target interface{}, deSeriMode DeSerializationMode, deSeriCtx interface{}, typeDen TypeDenotationType, serSel SerializableReadGuardFunc, errProducer ErrProducer) (*Deserializer, uint32) {
	if d.err != nil {
		return d, 0
	}
	ty, err := d.GetObjectType(typeDen)
	if err != nil {
		d.err = errProducer(err)

		return d, 0
	}
	seri, err := serSel(ty)
	if err != nil {
		d.err = errProducer(err)

		return d, 0
	}

	bytesConsumed, err := seri.Deserialize(d.src[d.offset:], deSeriMode, deSeriCtx)
	if err != nil {
		d.err = errProducer(err)

		return d, 0
	}

	d.offset += bytesConsumed
	d.readSerializableIntoTarget(target, seri)

	return d, ty
}

// ReadSliceOfObjects reads a slice of objects.
func (d *Deserializer) ReadSliceOfObjects(
	target interface{}, deSeriMode DeSerializationMode, deSeriCtx interface{}, lenType SeriLengthPrefixType,
	typeDen TypeDenotationType, arrayRules *ArrayRules, errProducer ErrProducer,
) *Deserializer {
	if d.err != nil {
		return d
	}

	var seris Serializables
	var seenTypes TypePrefixes
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		seenTypes = make(TypePrefixes, 0)
	}
	deserializeItem := func(b []byte) (bytesRead int, err error) {
		var seri Serializable

		// this mutates d.src/d.offset
		subDeseri := NewDeserializer(b)
		_, ty := subDeseri.readObject(func(readSeri Serializable) { seri = readSeri }, deSeriMode, deSeriCtx, typeDen, arrayRules.Guards.ReadGuard, func(err error) error {
			return errProducer(err)
		})

		bytesRead, err = subDeseri.Done()
		if err != nil {
			return 0, err
		}

		if deSeriMode.HasMode(DeSeriModePerformValidation) {
			seenTypes[ty] = struct{}{}
			if arrayRules.Guards.PostReadGuard != nil {
				if err := arrayRules.Guards.PostReadGuard(seri); err != nil {
					return 0, err
				}
			}
		}

		seris = append(seris, seri)

		return bytesRead, nil
	}

	d.ReadSequenceOfObjects(deserializeItem, deSeriMode, lenType, arrayRules, errProducer)
	if d.err != nil {
		return d
	}

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if !arrayRules.MustOccur.Subset(seenTypes) {
			d.err = errProducer(ierrors.Wrapf(ErrArrayValidationTypesNotOccurred, "should %v, has %v", arrayRules.MustOccur, seenTypes))

			return d
		}
	}

	if len(seris) == 0 {
		return d
	}

	d.readSerializablesIntoTarget(target, seris)

	return d
}

// DeserializeFunc is a function that reads bytes from b and returns how much bytes was read.
type DeserializeFunc func(b []byte) (bytesRead int, err error)

// ReadSequenceOfObjects reads a sequence of objects and calls DeserializeFunc for evey encountered item.
func (d *Deserializer) ReadSequenceOfObjects(
	itemDeserializer DeserializeFunc, deSeriMode DeSerializationMode,
	lenType SeriLengthPrefixType, arrayRules *ArrayRules, errProducer ErrProducer,
) *Deserializer {
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

	if sliceLength == 0 {
		return d
	}

	for i := 0; i < sliceLength; i++ {

		// Remember where we were before reading the item.
		srcBefore := d.src[d.offset:]
		offsetBefore := d.offset

		bytesRead, err := itemDeserializer(srcBefore)
		if err != nil {
			d.err = errProducer(err)

			return d
		}
		d.offset = offsetBefore + bytesRead

		if arrayElementValidator != nil {
			if err := arrayElementValidator(i, srcBefore[:bytesRead]); err != nil {
				d.err = errProducer(err)

				return d
			}
		}
	}

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

	remainingLen := len(d.src[d.offset:])
	if remainingLen < UInt64ByteSize {
		d.err = errProducer(ErrDeserializationNotEnoughData)

		return d
	}

	seconds := binary.LittleEndian.Uint64(d.src[d.offset : d.offset+UInt64ByteSize])

	// If the number of seconds in unit64 timestamp representation exceeds the max number of
	// seconds that can be represented in a int64 timestamp truncate to max of int64.
	if seconds > math.MaxInt64 {
		seconds = math.MaxInt64
	}

	*dest = time.Unix(int64(seconds), 0).UTC()

	d.offset += UInt64ByteSize

	return d
}

// ReadPayloadLength reads the payload length from the deserializer.
func (d *Deserializer) ReadPayloadLength() (uint32, error) {
	if len(d.src[d.offset:]) < PayloadLengthByteSize {
		return 0, ierrors.Wrap(ErrDeserializationNotEnoughData, "data is smaller than payload length denotation")
	}

	payloadLength := binary.LittleEndian.Uint32(d.src[d.offset:])
	d.offset += PayloadLengthByteSize

	return payloadLength, nil
}

// ReadPayload reads a payload.
func (d *Deserializer) ReadPayload(s interface{}, deSeriMode DeSerializationMode, deSeriCtx interface{}, sel SerializableReadGuardFunc, errProducer ErrProducer) *Deserializer {
	if d.err != nil {
		return d
	}
	payloadLength, err := d.ReadPayloadLength()
	if err != nil {
		d.err = errProducer(err)

		return d
	}

	// nothing to do
	if payloadLength == 0 {
		return d
	}

	switch {
	case len(d.src[d.offset:]) < MinPayloadByteSize:
		d.err = errProducer(ierrors.Wrapf(ErrDeserializationNotEnoughData, "payload data is smaller than min. required length %d", MinPayloadByteSize))

		return d
	case len(d.src[d.offset:]) < int(payloadLength):
		d.err = errProducer(ierrors.Wrap(ErrDeserializationNotEnoughData, "payload length denotes more bytes than are available"))

		return d
	}

	payload, err := sel(binary.LittleEndian.Uint32(d.src[d.offset:]))
	if err != nil {
		d.err = errProducer(err)

		return d
	}

	payloadBytesConsumed, err := payload.Deserialize(d.src[d.offset:], deSeriMode, deSeriCtx)
	if err != nil {
		d.err = errProducer(err)

		return d
	}

	if payloadBytesConsumed != int(payloadLength) {
		d.err = errProducer(ierrors.Wrapf(ErrInvalidBytes, "denoted payload length (%d) doesn't equal the size of deserialized payload (%d)", payloadLength, payloadBytesConsumed))

		return d
	}

	d.offset += payloadBytesConsumed
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
func (d *Deserializer) ReadString(s *string, lenType SeriLengthPrefixType, errProducer ErrProducer, minLen int, maxLen int) *Deserializer {
	if d.err != nil {
		return d
	}

	strLen, err := d.readSliceLength(lenType, errProducer)
	if err != nil {
		d.err = err

		return d
	}

	switch {
	case maxLen > 0 && strLen > maxLen:
		d.err = errProducer(ierrors.Wrapf(ErrDeserializationLengthInvalid, "string defined to be of %d bytes length but max %d is allowed", strLen, maxLen))
	case minLen > 0 && strLen < minLen:
		d.err = errProducer(ierrors.Wrapf(ErrDeserializationLengthInvalid, "string defined to be of %d bytes length but min %d is required", strLen, minLen))
	}

	if len(d.src[d.offset:]) < strLen {
		d.err = errProducer(ierrors.Wrapf(ErrDeserializationNotEnoughData, "data is smaller than (%d) denoted string length of %d", len(d.src[d.offset:]), strLen))

		return d
	}

	*s = string(d.src[d.offset : d.offset+strLen])

	d.offset += strLen

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

// WithValidation runs errProducer if deSeriMode has DeSeriModePerformValidation.
func (d *Deserializer) WithValidation(deSeriMode DeSerializationMode, errProducer ErrProducerWithRWBytes) *Deserializer {
	if d.err != nil {
		return d
	}
	if !deSeriMode.HasMode(DeSeriModePerformValidation) {
		return d
	}
	if err := errProducer(d.src[:d.offset], d.err); err != nil {
		d.err = err

		return d
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
		if err := CheckType(d.src[d.offset:], prefix); err != nil {
			d.err = errProducer(err)

			return d
		}
		toSkip = UInt32ByteSize
	case TypeDenotationByte:
		if err := CheckTypeByte(d.src[d.offset:], byte(prefix)); err != nil {
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

	if len(d.src) != d.offset {
		d.err = errProducer(len(d.src[d.offset:]), ErrDeserializationNotAllConsumed)
	}

	return d
}

// Done finishes the Deserializer by returning the read bytes and occurred errors.
func (d *Deserializer) Done() (int, error) {
	return d.offset, d.err
}
