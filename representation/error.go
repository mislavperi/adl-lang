package representation

type Error struct {
	Message string
}

func (e *Error) Type() RepresentationType { return ERROR_REPR }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
