package index

type Type interface {
	~int64
}

func Max[I Type](i, o I) I {
	if i > o {
		return i
	}
	return o
}

type IndexedID[I Type] interface {
	comparable

	Index() I
	String() string
}

type IndexedEntity[I Type, IndexedIDType IndexedID[I]] interface {
	ID() IndexedIDType
}
