package representation

type Environment struct {
	store map[string]Representation
	outer *Environment
}

func NewEnvironment() *Environment {
	s := make(map[string]Representation)
	return &Environment{store: s, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (Representation, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Representation) Representation {
	e.store[name] = val
	return val
}
