package representation

import "fmt"

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{Fn: func(args ...Representation) Representation {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
		},
	},
	{
		"out",
		&Builtin{

			Fn: func(args ...Representation) Representation {
				for _, arg := range args {
					fmt.Println(arg.Inspect())
				}
				return nil
			},
		},
	},
	{
		"first",
		&Builtin{

			Fn: func(args ...Representation) Representation {
				if len(args) != 1 {
					return newError("wrong number of arguments, got=%d, want=1", len(args))
				}
				if args[0].Type() != ARRAY_REPR {
					return newError("argument to `first` must be an array, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				if len(arr.Elements) > 0 {
					return arr.Elements[0]
				}

				return nil
			},
		},
	},
	{
		"last",
		&Builtin{

			Fn: func(args ...Representation) Representation {
				if len(args) != 1 {
					return newError("wrong number of arguments, got=%d, want=1", len(args))
				}
				if args[0].Type() != ARRAY_REPR {
					return newError("argument to `last` must be an array, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				if len(arr.Elements) > 0 {
					return arr.Elements[len(arr.Elements)-1]
				}

				return nil
			},
		},
	},
	{
		"rest",
		&Builtin{

			Fn: func(args ...Representation) Representation {
				if len(args) != 1 {
					return newError("wrong number of arguments, got=%d, want=1", len(args))
				}
				if args[0].Type() != ARRAY_REPR {
					return newError("argument to `rest` must be an array, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				length := len(arr.Elements)
				if length > 0 {
					restArray := make([]Representation, length-1)
					copy(restArray, arr.Elements[1:length])
					return &Array{Elements: restArray}
				}

				return nil
			},
		},
	},

	{
		"push",
		&Builtin{

			Fn: func(args ...Representation) Representation {
				if len(args) != 2 {
					return newError("wrong number of arguments, got=%d, want=2", len(args))
				}
				if args[0].Type() != ARRAY_REPR {
					return newError("argument to `push` must be an array, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				length := len(arr.Elements)

				newArray := make([]Representation, length+1)
				copy(newArray, arr.Elements)
				newArray[length] = args[1]

				return &Array{Elements: newArray}
			},
		},
	},
}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}
