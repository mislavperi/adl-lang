package representation

type RepresentationType string

const (
	INTEGER_REPR           RepresentationType = "INTEGER"
	BOOLEAN_REPR           RepresentationType = "BOOLEAN"
	NULL_REPR              RepresentationType = "NULL"
	RETURN_VALUE_REPR      RepresentationType = "RETURN_VALUE"
	ERROR_REPR             RepresentationType = "ERROR"
	FUNCTION_REPR          RepresentationType = "FUNCTION"
	STRING_REPR            RepresentationType = "STRING"
	BUILTIN_REPR           RepresentationType = "BUILTIN"
	ARRAY_REPR             RepresentationType = "ARRAY"
	HASH_REPR              RepresentationType = "HASH"
	COMPILED_FUNCTION_REPR RepresentationType = "COMPILED_FUNCTION"
	CLOSURE_REPR           RepresentationType = "CLOSURE"
)

type Representation interface {
	Type() RepresentationType
	Inspect() string
}
