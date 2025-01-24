package representation

import (
	"bytes"
	"strings"
)

type Array struct {
	Elements []Representation
}

func (ao *Array) Type() RepresentationType { return ARRAY_REPR }
func (ao *Array) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}
