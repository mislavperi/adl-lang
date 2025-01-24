package representation

import (
	"fmt"

	"github.com/mislavperi/adl-lang/code"
)

type CompiledFunction struct {
	Instructions  code.Instructions
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) Type() RepresentationType { return COMPILED_FUNCTION_REPR }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}
