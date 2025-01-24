package representation

import "fmt"

type Closure struct {
	Fn   *CompiledFunction
	Free []Representation
}

func (c *Closure) Type() RepresentationType { return CLOSURE_REPR }
func (c *Closure) Inspect() string {
	return fmt.Sprintf("Closure[%p]", c)
}
