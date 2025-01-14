package vm

import (
	"fmt"

	"github.com/mislavperi/adl-lang/code"
	"github.com/mislavperi/adl-lang/compiler"
	"github.com/mislavperi/adl-lang/object"
)

const GlobalsSize = 65536
const StackSize = 2048
const MaxFrames = 1024

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VM struct {
	constants []object.Object

	globals []object.Object

	stack        []object.Object
	stackPointer int

	frames      []*Frame
	framesIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,

		globals: make([]object.Object, GlobalsSize),

		stack:        make([]object.Object, StackSize),
		stackPointer: 0,

		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithGlobalStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) Run() error {
	var instructonPointer int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().instructonPointer < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().instructonPointer++

		instructonPointer = vm.currentFrame().instructonPointer
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[instructonPointer])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[instructonPointer+1:])
			vm.currentFrame().instructonPointer += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpGreaterThan, code.OpEqual, code.OpNotEqual:
			if err := vm.executeComparison(op); err != nil {
				return err
			}
		case code.OpBang:
			if err := vm.executeBangOperator(); err != nil {
				return err
			}
		case code.OpMinus:
			if err := vm.executeMinusOperator(); err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[instructonPointer+1:]))
			vm.currentFrame().instructonPointer = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[instructonPointer+1:]))
			vm.currentFrame().instructonPointer += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().instructonPointer = pos - 1
			}
		case code.OpNull:
			if err := vm.push(Null); err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[instructonPointer+1:])
			vm.currentFrame().instructonPointer += 2
			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[instructonPointer+1:])
			vm.currentFrame().instructonPointer += 2

			if err := vm.push(vm.globals[globalIndex]); err != nil {
				return err
			}
		case code.OpArray:
			numElements := int(code.ReadUint16(ins[instructonPointer+1:]))
			vm.currentFrame().instructonPointer += 2
			array := vm.buildArray(vm.stackPointer-numElements, vm.stackPointer)
			vm.stackPointer = vm.stackPointer - numElements
			if err := vm.push(array); err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUint16(ins[instructonPointer+1:]))
			vm.currentFrame().instructonPointer += 2

			hash, err := vm.buildHash(vm.stackPointer-numElements, vm.stackPointer)
			if err != nil {
				return err
			}
			vm.stackPointer = vm.stackPointer - numElements

			err = vm.push(hash)
			if err != nil {
				return nil
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			if err := vm.executeIndexExpression(left, index); err != nil {
				return err
			}
		case code.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.stackPointer = frame.basePointer - 1

			if err := vm.push(returnValue); err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			vm.stackPointer = frame.basePointer - 1

			if err := vm.push(Null); err != nil {
				return err
			}
		case code.OpCall:
			numArgs := code.ReadUint8(ins[instructonPointer+1:])
			vm.currentFrame().instructonPointer += 1

			if err := vm.executeCall(int(numArgs)); err != nil {
				return err
			}
		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[instructonPointer+1:])
			vm.currentFrame().instructonPointer += 1

			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()
		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[instructonPointer+1:])
			vm.currentFrame().instructonPointer += 1

			frame := vm.currentFrame()
			if err := vm.push(vm.stack[frame.basePointer+int(localIndex)]); err != nil {

				return err
			}
		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(ins[instructonPointer+1:])
			vm.currentFrame().instructonPointer += 1

			definition := object.Builtins[builtinIndex]

			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
		case code.OpClosure:
			constIndex := code.ReadUint16(ins[instructonPointer+1:])
			numFree := code.ReadUint8(ins[instructonPointer+3:])
			vm.currentFrame().instructonPointer += 3

			if err := vm.pushClosure(int(constIndex), int(numFree)); err != nil {
				return err
			}

		case code.OpGetFree:
			freeIndex := code.ReadUint8(ins[instructonPointer+1:])
			vm.currentFrame().instructonPointer += 1

			currentClosure := vm.currentFrame().closure
			err := vm.push(currentClosure.Free[freeIndex])
			if err != nil {
				return err
			}

		case code.OpCurrentClosure:
			currentClosure := vm.currentFrame().closure
			err := vm.push(currentClosure)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.stackPointer]
}

func (vm *VM) push(o object.Object) error {
	if vm.stackPointer >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.stackPointer] = o
	vm.stackPointer++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.stackPointer-1]
	vm.stackPointer--
	return o
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {

	case leftType == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(left, op, right)

	case leftType == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(left, op, right)
	default:
		return fmt.Errorf("unsupported type for binary operation: %s %s", leftType, rightType)
	}
}

func (vm *VM) executeBinaryIntegerOperation(left object.Object, operator code.Opcode, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64

	switch operator {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", operator)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(left object.Object, operator code.Opcode, right object.Object) error {
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	var result string

	switch operator {
	case code.OpAdd:
		result = leftValue + rightValue
	default:
		return fmt.Errorf("unknown string operator: %d", operator)
	}

	return vm.push(&object.String{Value: result})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(left, op, right)
	}
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)",
			op, left.Type(), right.Type())
	}

}

func (vm *VM) executeIntegerComparison(left object.Object, operator code.Opcode, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch operator {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", operator)
	}

}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func (vm *VM) executeIndexExpression(left object.Object, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)

	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array object.Object, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash object.Object, index object.Object) error {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) buildArray(startIndex int, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &object.Array{Elements: elements}

}

func (vm *VM) buildHash(startIndex int, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) callBuiltin(builtin *object.Builtin, argumentNumber int) error {
	args := vm.stack[vm.stackPointer-argumentNumber : vm.stackPointer]

	results := builtin.Fn(args...)
	vm.stackPointer = vm.stackPointer - argumentNumber - 1

	if results != nil {
		vm.push(results)
	} else {
		vm.push(Null)
	}

	return nil
}

func (vm *VM) callClosure(closure *object.Closure, argumentNumbers int) error {
	if argumentNumbers != closure.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			closure.Fn.NumParameters, argumentNumbers)
	}

	frame := NewFrame(closure, vm.stackPointer-argumentNumbers)
	vm.pushFrame(frame)

	vm.stackPointer = frame.basePointer + closure.Fn.NumLocals

	return nil
}

func (vm *VM) executeCall(argumentNumber int) error {
	calle := vm.stack[vm.stackPointer-1-argumentNumber]
	switch calle := calle.(type) {
	case *object.Closure:
		return vm.callClosure(calle, argumentNumber)
	case *object.Builtin:
		return vm.callBuiltin(calle, argumentNumber)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
}

func (vm *VM) pushClosure(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]
	fn, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.stackPointer-numFree+i]
	}
	vm.stackPointer = vm.stackPointer - numFree

	closure := &object.Closure{Fn: fn, Free: free}
	return vm.push(closure)
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}
