# Reflection Serializer

It is designed to serialize any objects to canonical and deterministic set of bytes.

## General principles

General principles of reflection serializer:

* integers and floats are little endian;
* if floats are NaN, error is returned during serialization;
* `time.Time` type is serialized as uint64 (unix nanoseconds). Zero value is serialized as 0;

* optionally, slices can be sorted in byte lexical order before serialization and deserialization;
* optionally, slices can have duplicates removed before serialization and deserialization;
* optionally, slices can be checked for duplicates during serialization/deserialization and return an error if 
  duplicates are found;
* optionally, slices can be checked if they are lexically sorted during serialization/deserialization and return an 
  error if the order is violated;
* by default, interface and pointer types cannot have `nil` value;
* by default, sizes of dynamic containers (`string`, `slice`) are written before values as uint32;
* optionally, collection types' length (e.g. `slice`, `orderedmap.OrderedMap`) can be validated before serialization and
  deserialization;
* golang's native `map` type is not supported;
* structs are serialized in the order of fields in the struct;
* only public fields can be serialized. To use private fields, read more about `unpack` struct tag;
* `encoding.BinaryUnmarshaler` and `encoding.BinaryMarshaler` interfaces are supported and used if available;
    * serialized bytes are serialized as a regular byte slice;
    * during deserialization byte slice is read and passed to unmarshaler;

* for custom collection types (e.g. `orderedmap.OrderedMap`), a special `BinarySerializer` and `BinaryDeserializer` is
  used. Custom collection types need to implement those interfaces internally;

## Available struct tags

Structure fields can be annotated using struct tags that can specify the following serialization details:

| Struct tag name | Type | Description                                                                                        | Supported types                          |
|-----------------|------|----------------------------------------------------------------------------------------------------|------------------------------------------|
| serialize       | bool | only fields with this tag will be serialized                                                       | all                                      |
| unpack          | bool | unpacks the inner structure's fields as fields of outer structure                                  | `struct`                                 |
| minLen          | int  | ensure that serialized collection has the specified min length.                                    | `slice` `orderedmap.OrderedMap`          |
| maxLen          | int  | ensure that serialized collection has the specified max length                                     | `slice` `orderedmap.OrderedMap`          |
| allowNil        | bool | whether given struct field can have nil value                                                      | `interface` `ptr`                        |
| noDuplicates    | bool | whether duplicate values are allowed in a slice. return an error if any duplicates are found       | `slice`                                  |
| skipDuplicates  | bool | whether a slice should have duplicate values removed during serialization and deserialization      | `slice`                                  |
| sort            | bool | whether a slice values should be lexicographically sorted before serialization and deserialization | `slice`                                  |
| lexicalOrder    | bool | whether to check if serialized/deserialized slices are sorted lexically. return an error otherwise | `slice`                                  |
| lenPrefixBytes  | int  | number of bytes to use for serialization of length of types with dynamic size.                     | `slice` `orderedmap.OrderedMap` `string` |

## Supported types

| Type      | How it's serialized |
|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| bool      | serialized as a single byte with value `0` or `1`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| int8      | serialized as a single byte                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| int16     | serialized as two bytes in Little Endian order                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| int32     | serialized as two bytes in Little Endian order                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| int64     | serialized as two bytes in Little Endian order                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| uint8     | serialized as a single byte                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| uint16    | serialized as two bytes in Little Endian order                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| uint32    | serialized as four bytes in Little Endian order                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| uint64    | serialized as eight bytes in Little Endian order                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| int       | serialized as `int64`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| uint      | serialized as `uint64`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| byte      | serialized as a single byte                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| float32   | serialized as four bytes in Little Endian order                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| float64   | serialized as eight bytes in Little Endian order                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| string    | serialized as slice of bytes in Go’s internal string encoding                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| array     | each element is serialized individually depending on the type. Length is contained in the type, so no need to add it as a prefix.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| slice     | each element is serialized individually depending on the type. Length is added before slice elements.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| map       | not supported. Use `orderedmap.OrderedMap` instead                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ptr       | * If pointer type implements built-in serialization interface, it is used to serialize its value. <br>* Otherwise, pointer value is unpacked and further serialized. <br>Furthermore, if pointer can be `nil` (according to struct tags), then the first byte is set either to indicate whether the value of the pointer is nil (0 for nil, 1 otherwise)                                                                                                                                                                                                                                                                                                                                                                                                                 |
| struct    | * If struct type implements built-in serialization interface, it is used to serialize the structure. <br>* If struct is `time.Time`, serialize nanosecond timestamp as `uint64` value. <br>* Otherwise, fields with `serialize:”true”` tag are serialized in the order of fields in the struct. Any other struct tags are picked up and used in field serialization.                                                                                                                                                                                                                                                                                                                                                                                                     |
| interface | Interface is only used to indicate that the value implements a set of methods, however for serialization, a concrete type needs to be known in order to serialize it and then deserialize it correctly. <br>First, underlying concrete type is resolved and using `TypeRegistry` it is translated into numeric value that is then serialized as `uint32`. During deserialization this value is used to correctly create concrete type value. <br>Second, when the type information is resolved then underlying concrete value is recursively serialized or deserialized. <br>Furthermore, if interface value can be `nil` (according to struct tags), then the first byte is set either to indicate whether the value of the interface field is nil (0 for nil, 1 otherwise) |

## Handling private fields

Reflection can read values from private fields, however it can only write values to public ones. That's why the
reflection serializer only allows serialization of public fields. However, using the fact that it can read private
values this limitation can be overcome by slightly modifying a struct structure. In the example below a structure has
one private field with pointer type. This would be impossible to deserialize such structure because reflection cannot
set values of private fields.

```go
package example

type SignatureUnlockBlock struct {
	signature *Signature
}
```

Instead, the following structure can be used to overcome the above problem. The structure will be serialized into the
same bytes, however during deserialization there will be no need to set the value of `signatureUnlockBlockInner`, as the
default zero value will be modified. The inner structure's serialized fields are public and their value can be set using
reflection. The `serialize:"unpack"` struct tag is necessary so that deserializer knows that it needs to use zero value
instead of setting the field value.

```go
package example

type SignatureUnlockBlock struct {
	signatureUnlockBlockInner `serialize:"unpack"`
}

type signatureUnlockBlockInner struct {
	Signature *Signature `serialize:"true"`
}
```

This approach solves the problem with serialization and requires just couple of lines of additional code to define outer
structure. The fields from inner structure will not be visible to the user as long as the names are shadowed by the
getter method with the same name, so direct access is still limited.

## Interface-type fields and `TypeRegistry`

For interface types or collection types that do not contain concrete types of the elements in the type definition, such
as slice of interface-type values or `orderedmap.OrderedMap` it is necessary to encode the concrete type information
into the serialized bytes. The concrete types are represented by a unique 4-byte `uint32` values. Which value
corresponds to a single concrete type in the code, so that deserializer knows which concrete type it needs to
instantiate for correct deserialization.

The relationships between numeric values and the types are stored by `TypeRegistry`. Each type that will be serialized
as a key or an element of `orderedmap.OrderedMap` or will be set as a value of field with interface type needs to be
added (registered) into the type registry. Registration will assign a unique integer value to the registered type. Type
registry makes it possible to transform the type into the numeric value and numeric value back into the concrete type.
Regular users will only need to register the types using type registry, however users that need to
implement `reflectionserializer.BinarySerializer` interface in their custom structures will also need to know how to
encode and decode the types.

### Registering new type

In order to register a new type, `RegisterType(value interface{}) error` needs to be called as in the example below. A
sample value of the registered type must be passed as an argument.

```go
sm := refseri.NewSerializer()
sm.RegisterType("") // register string type for use in orderedmap.OreredMap
sm.RegisterType(&ledgerstate.ED25519Address{}) // register concrete Address type
sm.RegisterType(&ledgerstate.ED25519Signature{}) // register concrete Signature type
sm.RegisterType(&ledgerstate.UTXOInput{}) // register concrete Input type
```

### Type encoding and decoding

In order to encode type, the `EncodeType(t reflect.Type) (uint32, error)` method needs to be called. For type decoding
similar `DecodeType(t uint32) (reflect.Type, error)` method needs to be called. The examples below show how it in
practice. The type registry is part of the serialization manager, so the `TypeRegistry` methods are available directly.
If a non-registered type is passed to either encoding or decoding function, it will result in an error that needs to be
handled. Unlike for type registration, for type encoding a reflection type must be passed and is returned from
the `TypeRegistry`.

```go
sm := refseri.NewSerializer()
sm.RegisterType("")

v := "stringType"
// encoding string type into corresponding integer value
encodedType, err := m.EncodeType(reflect.TypeOf(v))

// decoding integer value into string type
keyType, err := m.DecodeType(encodedType)
```

## Built-in serialization interfaces

### `encoding.BinaryMarshaller`

If struct or pointer type implements `encoding.BinaryMarshaller` and `encoding.BinaryUnmarshaller`, then the method is
used for serialization and deserialization. The interfaces are defined in standard Go libraries. The resulting bytes are
serialized as a regular slice of bytes, therefore the serialized value is prefixed with the number of bytes the
structure has been serialized to. During the serialization the length is used to read necessary number of bytes, and
then those bytes are passed to `UnmarshalBinary` method.

### `reflectionserializer.BinarySerializer`

If struct or pointer type implements `reflectionserializer.BinarySerializer`
and `reflectionserializer.BinaryDeserializer`, then the built-in method is used for serialization and deserialization.
The interface accepts buffer as one of the arguments, therefore there is no need to write number of serialized bytes.
This interface acts as an extension to serializer for complex structures such as `orderedmap.OrderedMap`, which can
serialize themselves using reflection serializer mechanisms. During deserialization the method reads the same buffer, so
there is no need to read all the bytes before passing them to the deserialization method. Below are the definitions of
the interfaces for reference:

```go
package reflectionserializer

import "github.com/iotaledger/hive.go/marshalutil"
import "github.com/iotaledger/hive.go/reflectionserializer"

// BinaryDeserializer interface is used to implement built-in deserialization of complex structures, usually collections.
type BinaryDeserializer interface {
	DeserializeBytes(buffer *marshalutil.MarshalUtil, m *reflectionserializer.SerializationManager, metadata reflectionserializer.FieldMetadata) (err error)
}

// BinarySerializer interface is used to implement built-in serialization of complex structures, usually collections.
type BinarySerializer interface {
	SerializeBytes(m *reflectionserializer.SerializationManager, metadata reflectionserializer.FieldMetadata) (data []byte, err error)
}
```

## Usage

Example below shows how the `TypeRegistry` can be used as well as how to serialize and deserialize slice containing
interface type.

```go
package example

import "github.com/iotaledger/hive.go/refseri"

type Test interface {
	Test() int
}

// *TestImpl1 implements Test
type TestImpl1 struct {
	Val int `serialize:"true"`
}

func (m TestImpl1) Test() int {
	return 1
}

// *TestImpl2 implements Test
type TestImpl2 struct {
	Val int `serialize:"true"`
}

func (m TestImpl2) Test() int {
	return 3
}

func InterfaceSlice() error {
	sm := refseri.NewSerializer()
	err := sm.RegisterType(TestImpl1{})
	if err != nil {
		return err
	}
	err = sm.RegisterType(TestImpl2{})
	if err != nil {
		return err
	}
	orig := []Test{TestImpl1{1}, TestImpl2{2}, TestImpl1{3}}
	bytes, err := sm.Serialize(orig)
	if err != nil {
		return err
	}
	var restored []Test
	err = sm.Deserialize(&restored, bytes)
	if err != nil {
		return err
	}
	return nil
}
```
