package objectstorage

type StorableObject interface {
	// updates the object "in place" (so it should use a pointer receiver)
	Update(other StorableObject)

	// returns the key that identifies the object
	GetStorageKey() []byte

	// returns the marshaled value that shall be stored in the database (without key)
	MarshalBinary() (data []byte, err error)

	// receives the concatenation of key and value as data (this way we can also "encode" things in the key and need
	// less space)
	UnmarshalBinary(data []byte) error
}
