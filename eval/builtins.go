package eval

import (
	"fmt"

	"github.com/mislavperi/adl-lang/representation"
)

var builtins = map[string]*representation.Builtin{
	"len": {
		Fn: func(args ...representation.Representation) representation.Representation {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *representation.Array:
				return &representation.Integer{Value: int64(len(arg.Elements))}
			case *representation.String:
				return &representation.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"first": {
		Fn: func(args ...representation.Representation) representation.Representation {
			if len(args) != 1 {
				return newError("wrong number of arguments, got=%d, want=1", len(args))
			}
			if args[0].Type() != representation.ARRAY_REPR {
				return newError("argument to `first` must be an ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*representation.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return NULL
		},
	},
	"last": {
		Fn: func(args ...representation.Representation) representation.Representation {
			if len(args) != 1 {
				return newError("wrong number of arguments, got=%d, want=1", len(args))
			}
			if args[0].Type() != representation.ARRAY_REPR {
				return newError("argument to `last` must be an ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*representation.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[len(arr.Elements)-1]
			}

			return NULL
		},
	},
	"rest": {
		Fn: func(args ...representation.Representation) representation.Representation {
			if len(args) != 2 {
				return newError("wrong number of arguments, got=%d, want=2", len(args))
			}
			if args[0].Type() != representation.ARRAY_REPR {
				return newError("argument to `rest` must be an ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*representation.Array)
			length := len(arr.Elements)
			if length > 0 {
				restArray := make([]representation.Representation, length-1)
				copy(restArray, arr.Elements[1:length])
				return &representation.Array{Elements: restArray}
			}

			return NULL
		},
	},
	"push": {
		Fn: func(args ...representation.Representation) representation.Representation {
			if len(args) != 1 {
				return newError("wrong number of arguments, got=%d, want=1", len(args))
			}
			if args[0].Type() != representation.ARRAY_REPR {
				return newError("argument to `rest` must be an ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*representation.Array)
			length := len(arr.Elements)

			newArray := make([]representation.Representation, length+1)
			copy(newArray, arr.Elements)
			newArray[length] = args[1]

			return &representation.Array{Elements: newArray}
		},
	},
	"out": {
		Fn: func(args ...representation.Representation) representation.Representation {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
}
