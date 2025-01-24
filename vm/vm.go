package vm

import (
	"fmt"

	"github.com/mislavperi/adl-lang/code"
	"github.com/mislavperi/adl-lang/compiler"
	"github.com/mislavperi/adl-lang/representation"
)

const GlobalsSize = 65536
const StackSize = 2048
const MaxFrames = 1024

var True = &representation.Boolean{Value: true}
var False = &representation.Boolean{Value: false}
var Null = &representation.Null{}

type VM struct {
	constants []representation.Representation

	globals []representation.Representation

	stack        []representation.Representation
	stackPointer int

	frames      []*Frame
	framesIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &representation.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &representation.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,

		globals: make([]representation.Representation, GlobalsSize),

		stack:        make([]representation.Representation, StackSize),
		stackPointer: 0,

		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithGlobalStore(bytecode *compiler.Bytecode, s []representation.Representation) *VM {
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

			definition := representation.Builtins[builtinIndex]

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

func (vm *VM) LastPoppedStackElem() representation.Representation {
	return vm.stack[vm.stackPointer]
}

func (vm *VM) push(o representation.Representation) error {
	if vm.stackPointer >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.stackPointer] = o
	vm.stackPointer++

	return nil
}

func (vm *VM) pop() representation.Representation {
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

	case leftType == representation.INTEGER_REPR && right.Type() == representation.INTEGER_REPR:
		return vm.executeBinaryIntegerOperation(left, op, right)

	case leftType == representation.STRING_REPR && right.Type() == representation.STRING_REPR:
		return vm.executeBinaryStringOperation(left, op, right)
	default:
		return fmt.Errorf("unsupported type for binary operation: %s %s", leftType, rightType)
	}
}

func (vm *VM) executeBinaryIntegerOperation(left representation.Representation, operator code.Opcode, right representation.Representation) error {
	leftValue := left.(*representation.Integer).Value
	rightValue := right.(*representation.Integer).Value

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

	return vm.push(&representation.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(left representation.Representation, operator code.Opcode, right representation.Representation) error {
	leftValue := left.(*representation.String).Value
	rightValue := right.(*representation.String).Value

	var result string

	switch operator {
	case code.OpAdd:
		result = leftValue + rightValue
	default:
		return fmt.Errorf("unknown string operator: %d", operator)
	}

	return vm.push(&representation.String{Value: result})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == representation.INTEGER_REPR && right.Type() == representation.INTEGER_REPR {
		return vm.executeIntegerComparison(left, op, right)
	}
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanrepresentation(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanrepresentation(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)",
			op, left.Type(), right.Type())
	}

}

func (vm *VM) executeIntegerComparison(left representation.Representation, operator code.Opcode, right representation.Representation) error {
	leftValue := left.(*representation.Integer).Value
	rightValue := right.(*representation.Integer).Value

	switch operator {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanrepresentation(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanrepresentation(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanrepresentation(leftValue > rightValue))
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

	if operand.Type() != representation.INTEGER_REPR {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*representation.Integer).Value
	return vm.push(&representation.Integer{Value: -value})
}

func (vm *VM) executeIndexExpression(left representation.Representation, index representation.Representation) error {
	switch {
	case left.Type() == representation.ARRAY_REPR && index.Type() == representation.INTEGER_REPR:
		return vm.executeArrayIndex(left, index)

	case left.Type() == representation.HASH_REPR:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array representation.Representation, index representation.Representation) error {
	arrayrepresentation := array.(*representation.Array)
	i := index.(*representation.Integer).Value
	max := int64(len(arrayrepresentation.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayrepresentation.Elements[i])
}

func (vm *VM) executeHashIndex(hash representation.Representation, index representation.Representation) error {
	hashrepresentation := hash.(*representation.Hash)

	key, ok := index.(representation.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashrepresentation.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) buildArray(startIndex int, endIndex int) representation.Representation {
	elements := make([]representation.Representation, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &representation.Array{Elements: elements}

}

func (vm *VM) buildHash(startIndex int, endIndex int) (representation.Representation, error) {
	hashedPairs := make(map[representation.HashKey]representation.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := representation.HashPair{Key: key, Value: value}
		hashKey, ok := key.(representation.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &representation.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) callBuiltin(builtin *representation.Builtin, argumentNumber int) error {
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

func (vm *VM) callClosure(closure *representation.Closure, argumentNumbers int) error {
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
	case *representation.Closure:
		return vm.callClosure(calle, argumentNumber)
	case *representation.Builtin:
		return vm.callBuiltin(calle, argumentNumber)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
}

func (vm *VM) pushClosure(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]
	fn, ok := constant.(*representation.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]representation.Representation, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.stackPointer-numFree+i]
	}
	vm.stackPointer = vm.stackPointer - numFree

	closure := &representation.Closure{Fn: fn, Free: free}
	return vm.push(closure)
}

func nativeBoolToBooleanrepresentation(input bool) *representation.Boolean {
	if input {
		return True
	}
	return False
}

func isTruthy(obj representation.Representation) bool {
	switch obj := obj.(type) {
	case *representation.Boolean:
		return obj.Value
	case *representation.Null:
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
