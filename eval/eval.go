package eval

import (
	"fmt"

	"github.com/mislavperi/adl-lang/ast"
	"github.com/mislavperi/adl-lang/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.ERROR_OBJ
}

func Evaluate(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		var result object.Object
		for _, statement := range node.Statements {
			result = Evaluate(statement, env)
			if result != nil {
				if returnValue, ok := result.(*object.ReturnValue); ok {
					return returnValue.Value
				}
				if result.Type() == object.ERROR_OBJ {
					return result
				}
			}
		}
		return result

	case *ast.ExpressionStatement:
		return Evaluate(node.Expression, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

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
			if right.Type() != object.INTEGER_OBJ {
				return newError("unknown operator: -%s", right.Type())
			}

			value := right.(*object.Integer).Value
			return &object.Integer{Value: -value}
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
		case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
			leftVal := left.(*object.Integer).Value
			rightVal := right.(*object.Integer).Value

			switch node.Operator {
			case "+":
				return &object.Integer{Value: leftVal + rightVal}
			case "-":
				return &object.Integer{Value: leftVal - rightVal}
			case "*":
				return &object.Integer{Value: leftVal * rightVal}
			case "/":
				return &object.Integer{Value: leftVal / rightVal}
			case "<":
				return booleanToBooleanObject(leftVal < rightVal)
			case ">":
				return booleanToBooleanObject(leftVal > rightVal)
			case "==":
				return booleanToBooleanObject(leftVal == rightVal)
			case "!=":
				return booleanToBooleanObject(leftVal != rightVal)
			default:
				return newError("unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
			}

		case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
			if node.Operator != "+" {
				return newError("unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
			}

			leftVal := left.(*object.String).Value
			rightVal := right.(*object.String).Value
			return &object.String{Value: leftVal + rightVal}

		case node.Operator == "==":
			return booleanToBooleanObject(left == right)
		case node.Operator == "!=":
			return booleanToBooleanObject(left != right)
		default:
			return newError("unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
		}

	case *ast.BlockStatement:
		var result object.Object

		for _, statement := range node.Statements {
			result = Evaluate(statement, env)
			if result != nil {
				if result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ {
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
		return &object.ReturnValue{Value: val}

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
		return &object.Function{Parameters: parameters, Env: env, Body: body}

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
		case *object.Function:
			extendedEnv := object.NewEnclosedEnvironment(fn.Env)
			for paramIdx, param := range fn.Parameters {
				extendedEnv.Set(param.Value, args[paramIdx])
			}
			evaluated := Evaluate(fn.Body, extendedEnv)
			if returnValue, ok := evaluated.(*object.ReturnValue); ok {
				return returnValue.Value
			}
			return evaluated
		case *object.Builtin:
			return fn.Fn(args...)
		default:
			return newError("not a function: %s", fn.Type())
		}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.ArrayLiteral:
		elements := evaluateExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

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
		case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
			arrayObject := left.(*object.Array)
			idx := index.(*object.Integer).Value
			max := int64(len(arrayObject.Elements) - 1)
			if idx < 0 || idx > max {
				return NULL
			}
			return arrayObject.Elements[idx]

		case left.Type() == object.HASH_OBJ:
			hashObject := left.(*object.Hash)
			key, ok := index.(object.Hashable)
			if !ok {
				return newError("unusable as hash key: %s", index.Type())
			}
			if pair, ok := hashObject.Pairs[key.HashKey()]; ok {
				return pair.Value
			}
			return NULL

		default:
			return newError("index operator not supported: %s", left.Type())
		}

	case *ast.HashLiteral:
		pairs := make(map[object.HashKey]object.HashPair)

		for keyNode, valueNode := range node.Pairs {
			key := Evaluate(keyNode, env)
			if isError(key) {
				return key
			}

			hashKey, ok := key.(object.Hashable)
			if !ok {
				return newError("unusable as a hash key:  %s", key.Type())
			}

			value := Evaluate(valueNode, env)
			if isError(value) {
				return value
			}

			hashed := hashKey.HashKey()
			pairs[hashed] = object.HashPair{Key: key, Value: value}
		}

		return &object.Hash{Pairs: pairs}
	}

	return nil
}

func booleanToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evaluateExpressions(expressions []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, expr := range expressions {
		evaluated := Evaluate(expr, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func isTruthy(obj object.Object) bool {
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

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}
