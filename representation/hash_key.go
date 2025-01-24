package representation

type HashKey struct {
	Type  RepresentationType
	Value uint64
}

type Hashable interface {
	HashKey() HashKey
}
