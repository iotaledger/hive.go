package index

type Type interface {
	~int64
}

type IndexedID[I Type] interface {
	comparable

	Index() I
	String() string
}

type IndexedEntity[I Type, IndexedIDType IndexedID[I]] interface {
	ID() IndexedIDType
}
