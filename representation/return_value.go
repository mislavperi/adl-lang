package representation

type ReturnValue struct {
	Value Representation
}

func (rv *ReturnValue) Type() RepresentationType { return RETURN_VALUE_REPR }
func (rv *ReturnValue) Inspect() string          { return rv.Value.Inspect() }
