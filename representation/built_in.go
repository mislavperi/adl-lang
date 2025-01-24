package representation

type BuiltinFunction func(args ...Representation) Representation

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() RepresentationType { return BUILTIN_REPR }
func (b *Builtin) Inspect() string          { return "builtin function" }
