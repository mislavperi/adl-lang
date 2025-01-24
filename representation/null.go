package representation

type Null struct{}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() RepresentationType { return NULL_REPR }
