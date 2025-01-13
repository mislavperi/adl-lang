package vm

import (
	"github.com/mislavperi/gem-lang/code"
	"github.com/mislavperi/gem-lang/object"
)

type Frame struct {
	closure           *object.Closure
	instructonPointer int
	basePointer       int
}

func NewFrame(closure *object.Closure, basePointer int) *Frame {
	return &Frame{closure: closure, instructonPointer: -1, basePointer: basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.closure.Fn.Instructions
}
