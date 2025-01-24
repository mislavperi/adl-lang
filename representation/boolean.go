package representation

import "fmt"

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string      { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() RepresentationType { return BOOLEAN_REPR }
func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	}
	return HashKey{Type: b.Type(), Value: value}
}
