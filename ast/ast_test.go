package ast

import (
	"testing"

	"github.com/mislavperi/gem-lang/token"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&LetStatement{
				BaseNode: BaseNode{Token: token.Token{Type: token.LET, Literal: "let"}},
				Name: &Identifier{
					BaseNode: BaseNode{Token: token.Token{Type: token.IDENTIFER, Literal: "myVar"}},
					Value:    "myVar",
				},
				Value: &Identifier{
					BaseNode: BaseNode{Token: token.Token{Type: token.IDENTIFER, Literal: "anotherVar"}},
					Value:    "anotherVar",
				},
			},
		},
	}

	if program.String() != "let myVar = anotherVar;" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
