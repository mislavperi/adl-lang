package object

type ObjectType string

const (
	INTEGER_OBJ           ObjectType = "INTEGER"
	BOOLEAN_OBJ           ObjectType = "BOOLEAN"
	NULL_OBJ              ObjectType = "NULL"
	RETURN_VALUE_OBJ      ObjectType = "RETURN_VALUE"
	ERROR_OBJ             ObjectType = "ERROR"
	FUNCTION_OBJ          ObjectType = "FUNCTION"
	STRING_OBJ            ObjectType = "STRING"
	BUILTIN_OBJ           ObjectType = "BUILTIN"
	ARRAY_OBJ             ObjectType = "ARRAY"
	HASH_OBJ              ObjectType = "HASH"
	COMPILED_FUNCTION_OBJ ObjectType = "COMPILED_FUNCTION"
	CLOSURE_OBJ           ObjectType = "CLOSURE"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}
