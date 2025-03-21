package representation

import (
	"hash/fnv"
)

type String struct {
	Value string
}

func (s *String) Type() RepresentationType { return STRING_REPR }
func (s *String) Inspect() string  { return s.Value }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}
