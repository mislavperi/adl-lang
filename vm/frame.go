package vm

import (
	"github.com/mislavperi/adl-lang/code"
	"github.com/mislavperi/adl-lang/representation"
)

type Frame struct {
	closure           *representation.Closure
	instructonPointer int
	basePointer       int
}

func NewFrame(closure *representation.Closure, basePointer int) *Frame {
	return &Frame{closure: closure, instructonPointer: -1, basePointer: basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.closure.Fn.Instructions
}
