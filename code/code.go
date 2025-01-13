package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

// String presents the Instructions in a readable format.
func (ins Instructions) String() string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.formatInstruction(def, operands))

		i += 1 + read
	}
	return out.String()
}

// formatInstruction formats the instruction with its operands.
func (ins Instructions) formatInstruction(def *Definition, operands []int) string {
	if len(operands) != len(def.OperandWidths) {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), len(def.OperandWidths))
	}

	switch len(def.OperandWidths) {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

type Opcode byte

const (
	OpConstant Opcode = iota
	OpAdd
	OpPop
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpMinus
	OpBang
	OpJumpNotTruthy
	OpJump
	OpNull
	OpGetGlobal
	OpSetGlobal
	OpArray
	OpHash
	OpIndex
	OpCall
	OpReturnValue
	OpReturn
	OpGetLocal
	OpSetLocal
	OpGetBuiltin
	OpClosure
	OpCurrentClosure
	OpGetFree
)

type Definition struct {
	Name          string
	OperandWidths []int
}

// Global variable for opcode definitions.
var definitions = map[Opcode]*Definition{
	OpConstant:       {"OpConstant", []int{2}},
	OpAdd:            {"OpAdd", []int{}},
	OpSub:            {"OpSub", []int{}},
	OpMul:            {"OpMul", []int{}},
	OpDiv:            {"OpDiv", []int{}},
	OpPop:            {"OpPop", []int{}},
	OpTrue:           {"OpTrue", []int{}},
	OpFalse:          {"OpFalse", []int{}},
	OpEqual:          {"OpEqual", []int{}},
	OpNotEqual:       {"OpNotEqual", []int{}},
	OpGreaterThan:    {"OpGreaterThan", []int{}},
	OpMinus:          {"OpMinus", []int{}},
	OpBang:           {"OpBang", []int{}},
	OpJumpNotTruthy:  {"OpJumpNotTruthy", []int{2}},
	OpJump:           {"OpJump", []int{2}},
	OpNull:           {"OpNull", []int{}},
	OpGetGlobal:      {"OpGetGlobal", []int{2}},
	OpSetGlobal:      {"OpSetGlobal", []int{2}},
	OpArray:          {"OpArray", []int{2}},
	OpHash:           {"OpHash", []int{2}},
	OpIndex:          {"OpIndex", []int{}},
	OpCall:           {"OpCall", []int{1}},
	OpReturnValue:    {"OpReturnValue", []int{}},
	OpReturn:         {"OpReturn", []int{}},
	OpGetLocal:       {"OpGetLocal", []int{1}},
	OpSetLocal:       {"OpSetLocal", []int{1}},
	OpGetBuiltin:     {"OpGetBuiltin", []int{1}},
	OpClosure:        {"OpClosure", []int{2, 1}},
	OpCurrentClosure: {"OpCurrentClosure", []int{}},
	OpGetFree:        {"OpGetFree", []int{1}},
}

// Lookup finds the definition for a given opcode.
func Lookup(op byte) (*Definition, error) {
	if def, ok := definitions[Opcode(op)]; ok {
		return def, nil
	}
	return nil, fmt.Errorf("opcode %d undefined", op)
}

// Make creates an instruction from an opcode and its operands.
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionsLen := 1 + sum(def.OperandWidths)
	instruction := make([]byte, instructionsLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

// ReadOperands reads the operands for a given instruction definition.
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}

	return operands, offset
}

// ReadUint16 reads a uint16 value from the instruction sequence.
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

// ReadUint8 reads a uint8 value from the instruction sequence.
func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

// sum calculates the sum of a slice of integers.
func sum(slice []int) int {
	total := 0
	for _, value := range slice {
		total += value
	}
	return total
}
