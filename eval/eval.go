package eval

import (
	"fmt"

	"github.com/mislavperi/adl-lang/ast"
	"github.com/mislavperi/adl-lang/representation"
)

var (
	NULL  = &representation.Null{}
	TRUE  = &representation.Boolean{Value: true}
	FALSE = &representation.Boolean{Value: false}
)

func isError(obj representation.Representation) bool {
	return obj != nil && obj.Type() == representation.ERROR_REPR
}

func Evaluate(node ast.Node, env *representation.Environment) representation.Representation {
	switch node := node.(type) {
	case *ast.Program:
		var result representation.Representation
		for _, statement := range node.Statements {
			result = Evaluate(statement, env)
			if result != nil {
				if returnValue, ok := result.(*representation.ReturnValue); ok {
					return returnValue.Value
				}
				if result.Type() == representation.ERROR_REPR {
					return result
				}
			}
		}
		return result

	case *ast.ExpressionStatement:
		return Evaluate(node.Expression, env)

	case *ast.IntegerLiteral:
		return &representation.Integer{Value: node.Value}

	case *ast.Boolean:
		if node.Value {
			return TRUE
		}
		return FALSE

	case *ast.PrefixExpression:
		right := Evaluate(node.Right, env)
		if isError(right) {
			return right
		}

		switch node.Operator {
		case "!":
			switch right {
			case TRUE:
				return FALSE
			case FALSE:
				return TRUE
			case NULL:
				return TRUE
			default:
				return FALSE
			}
		case "-":
			if right.Type() != representation.INTEGER_REPR {
				return newError("unknown operator: -%s", right.Type())
			}

			value := right.(*representation.Integer).Value
			return &representation.Integer{Value: -value}
		default:
			return newError("unknown operator: %s%s", node.Operator, right.Type())
		}

	case *ast.InfixExpression:
		left := Evaluate(node.Left, env)
		if isError(left) {
			return left
		}

		right := Evaluate(node.Right, env)
		if isError(right) {
			return right
		}

		if left.Type() != right.Type() {
			return newError("type mismatch: %s %s %s", left.Type(), node.Operator, right.Type())
		}

		switch {
		case left.Type() == representation.INTEGER_REPR && right.Type() == representation.INTEGER_REPR:
			leftVal := left.(*representation.Integer).Value
			rightVal := right.(*representation.Integer).Value

			switch node.Operator {
			case "+":
				return &representation.Integer{Value: leftVal + rightVal}
			case "-":
				return &representation.Integer{Value: leftVal - rightVal}
			case "*":
				return &representation.Integer{Value: leftVal * rightVal}
			case "/":
				return &representation.Integer{Value: leftVal / rightVal}
			case "<":
				return booleanToBooleanRepresentation(leftVal < rightVal)
			case ">":
				return booleanToBooleanRepresentation(leftVal > rightVal)
			case "==":
				return booleanToBooleanRepresentation(leftVal == rightVal)
			case "!=":
				return booleanToBooleanRepresentation(leftVal != rightVal)
			default:
				return newError("unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
			}

		case left.Type() == representation.STRING_REPR && right.Type() == representation.STRING_REPR:
			if node.Operator != "+" {
				return newError("unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
			}

			leftVal := left.(*representation.String).Value
			rightVal := right.(*representation.String).Value
			return &representation.String{Value: leftVal + rightVal}

		case node.Operator == "==":
			return booleanToBooleanRepresentation(left == right)
		case node.Operator == "!=":
			return booleanToBooleanRepresentation(left != right)
		default:
			return newError("unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
		}

	case *ast.BlockStatement:
		var result representation.Representation

		for _, statement := range node.Statements {
			result = Evaluate(statement, env)
			if result != nil {
				if result.Type() == representation.RETURN_VALUE_REPR || result.Type() == representation.ERROR_REPR {
					return result
				}
			}
		}

		return result

	case *ast.IfExpression:
		condition := Evaluate(node.Condition, env)
		if isError(condition) {
			return condition
		}

		if isTruthy(condition) {
			return Evaluate(node.Consequence, env)
		} else if node.Alternative != nil {
			return Evaluate(node.Alternative, env)
		} else {
			return NULL
		}

	case *ast.ReturnStatement:
		val := Evaluate(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &representation.ReturnValue{Value: val}

	case *ast.LetStatement:
		val := Evaluate(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return nil

	case *ast.Identifier:
		if val, ok := env.Get(node.Value); ok {
			return val
		}

		if builtin, ok := builtins[node.Value]; ok {
			return builtin
		}

		return newError("identifier not found: " + node.Value)

	case *ast.FnLiteral:
		parameters := node.Parameters
		body := node.Body
		return &representation.Function{Parameters: parameters, Env: env, Body: body}

	case *ast.CallExpression:
		function := Evaluate(node.Function, env)
		if isError(function) {
			return function
		}

		args := evaluateExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		switch fn := function.(type) {
		case *representation.Function:
			extendedEnv := representation.NewEnclosedEnvironment(fn.Env)
			for paramIdx, param := range fn.Parameters {
				extendedEnv.Set(param.Value, args[paramIdx])
			}
			evaluated := Evaluate(fn.Body, extendedEnv)
			if returnValue, ok := evaluated.(*representation.ReturnValue); ok {
				return returnValue.Value
			}
			return evaluated
		case *representation.Builtin:
			return fn.Fn(args...)
		default:
			return newError("not a function: %s", fn.Type())
		}

	case *ast.StringLiteral:
		return &representation.String{Value: node.Value}

	case *ast.ArrayLiteral:
		elements := evaluateExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &representation.Array{Elements: elements}

	case *ast.IndexExpression:
		left := Evaluate(node.Left, env)
		if isError(left) {
			return left
		}
		index := Evaluate(node.Index, env)
		if isError(index) {
			return index
		}

		switch {
		case left.Type() == representation.ARRAY_REPR && index.Type() == representation.INTEGER_REPR:
			arrayRepresentation := left.(*representation.Array)
			idx := index.(*representation.Integer).Value
			max := int64(len(arrayRepresentation.Elements) - 1)
			if idx < 0 || idx > max {
				return NULL
			}
			return arrayRepresentation.Elements[idx]

		case left.Type() == representation.HASH_REPR:
			hashRepresentation := left.(*representation.Hash)
			key, ok := index.(representation.Hashable)
			if !ok {
				return newError("unusable as hash key: %s", index.Type())
			}
			if pair, ok := hashRepresentation.Pairs[key.HashKey()]; ok {
				return pair.Value
			}
			return NULL

		default:
			return newError("index operator not supported: %s", left.Type())
		}

	case *ast.HashLiteral:
		pairs := make(map[representation.HashKey]representation.HashPair)

		for keyNode, valueNode := range node.Pairs {
			key := Evaluate(keyNode, env)
			if isError(key) {
				return key
			}

			hashKey, ok := key.(representation.Hashable)
			if !ok {
				return newError("unusable as a hash key:  %s", key.Type())
			}

			value := Evaluate(valueNode, env)
			if isError(value) {
				return value
			}

			hashed := hashKey.HashKey()
			pairs[hashed] = representation.HashPair{Key: key, Value: value}
		}

		return &representation.Hash{Pairs: pairs}
	}

	return nil
}

func booleanToBooleanRepresentation(input bool) *representation.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evaluateExpressions(expressions []ast.Expression, env *representation.Environment) []representation.Representation {
	var result []representation.Representation

	for _, expr := range expressions {
		evaluated := Evaluate(expr, env)
		if isError(evaluated) {
			return []representation.Representation{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func isTruthy(obj representation.Representation) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *representation.Error {
	return &representation.Error{Message: fmt.Sprintf(format, a...)}
}
