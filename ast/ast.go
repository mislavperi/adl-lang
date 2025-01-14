package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mislavperi/adl-lang/token"
)

// Node represents a node in the AST and provides methods for token interaction and string representation.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement is an interface for AST statement nodes.
type Statement interface {
	Node
	isStatement()
}

// Expression is an interface for AST expression nodes.
type Expression interface {
	Node
	isExpression()
}

// BaseNode is an embedded struct to reduce code duplication for TokenLiteral and String methods.
type BaseNode struct {
	Token token.Token
}

func (b *BaseNode) TokenLiteral() string { return b.Token.Literal }

// Identifier represents an identifier node in the AST.
type Identifier struct {
	BaseNode
	Value string
}

func (i *Identifier) isExpression()  {}
func (i *Identifier) String() string { return i.Value }

// Program is the root node of AST, containing a set of statements.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var output bytes.Buffer
	for _, stmt := range p.Statements {
		output.WriteString(stmt.String())
	}
	return output.String()
}

// LetStatement represents a let statement node in the AST.
type LetStatement struct {
	BaseNode
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) isStatement() {}
func (ls *LetStatement) String() string {
	var output bytes.Buffer
	output.WriteString(fmt.Sprintf("%s %s = %s;", ls.TokenLiteral(), ls.Name.String(), ls.Value.String()))
	return output.String()
}

// ReturnStatement represents a return statement node in the AST.
type ReturnStatement struct {
	BaseNode
	ReturnValue Expression
}

func (rs *ReturnStatement) isStatement() {}
func (rs *ReturnStatement) String() string {
	var output bytes.Buffer
	output.WriteString(rs.TokenLiteral())
	if rs.ReturnValue != nil {
		output.WriteString(" " + rs.ReturnValue.String())
	}
	output.WriteString(";")
	return output.String()
}

// ExpressionStatement represents a standalone expression node in the AST.
type ExpressionStatement struct {
	BaseNode
	Expression Expression
}

func (es *ExpressionStatement) isStatement() {}
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// IntegerLiteral represents an integer literal node in the AST.
type IntegerLiteral struct {
	BaseNode
	Value int64
}

func (il *IntegerLiteral) isExpression()  {}
func (il *IntegerLiteral) String() string { return il.Token.Literal }

// PrefixExpression represents a prefix expression node in the AST.
type PrefixExpression struct {
	BaseNode
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) isExpression() {}
func (pe *PrefixExpression) String() string {
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String())
}

// InfixExpression represents an infix expression node in the AST.
type InfixExpression struct {
	BaseNode
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) isExpression() {}
func (ie *InfixExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", ie.Left.String(), ie.Operator, ie.Right.String())
}

// Boolean represents a boolean node in the AST.
type Boolean struct {
	BaseNode
	Value bool
}

func (b *Boolean) isExpression()  {}
func (b *Boolean) String() string { return b.Token.Literal }

// BlockStatement is a series of statements enclosed in a block in the AST.
type BlockStatement struct {
	BaseNode
	Statements []Statement
}

func (bs *BlockStatement) isStatement() {}
func (bs *BlockStatement) String() string {
	var output bytes.Buffer
	for _, stmt := range bs.Statements {
		output.WriteString(stmt.String())
	}
	return output.String()
}

// IfExpression represents an if expression node in the AST.
type IfExpression struct {
	BaseNode
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) isExpression() {}
func (ie *IfExpression) String() string {
	var output bytes.Buffer
	output.WriteString(fmt.Sprintf("if %s %s", ie.Condition.String(), ie.Consequence.String()))
	if ie.Alternative != nil {
		output.WriteString(" else " + ie.Alternative.String())
	}
	return output.String()
}

// FnLiteral represents a function literal node in the AST.
type FnLiteral struct {
	BaseNode
	Parameters []*Identifier
	Body       *BlockStatement
	Name       string
}

func (fl *FnLiteral) isExpression() {}
func (fl *FnLiteral) String() string {
	params := make([]string, len(fl.Parameters))
	for i, param := range fl.Parameters {
		params[i] = param.String()
	}
	if fl.Name != "" {
		return fmt.Sprintf("%s<%s>(%s) %s", fl.TokenLiteral(), fl.Name, strings.Join(params, ", "), fl.Body.String())
	}
	return fmt.Sprintf("%s(%s) %s", fl.TokenLiteral(), strings.Join(params, ", "), fl.Body.String())
}

// CallExpression represents a call expression in the AST.
type CallExpression struct {
	BaseNode
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) isExpression() {}
func (ce *CallExpression) String() string {
	args := make([]string, len(ce.Arguments))
	for i, arg := range ce.Arguments {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", ce.Function.String(), strings.Join(args, ", "))
}

// StringLiteral represents a string literal in the AST.
type StringLiteral struct {
	BaseNode
	Value string
}

func (sl *StringLiteral) isExpression()  {}
func (sl *StringLiteral) String() string { return sl.Token.Literal }

// ArrayLiteral represents an array literal in the AST.
type ArrayLiteral struct {
	BaseNode
	Elements []Expression
}

func (al *ArrayLiteral) isExpression() {}
func (al *ArrayLiteral) String() string {
	elements := make([]string, len(al.Elements))
	for i, el := range al.Elements {
		elements[i] = el.String()
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

// IndexExpression represents an index operation in the AST.
type IndexExpression struct {
	BaseNode
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) isExpression() {}
func (ie *IndexExpression) String() string {
	return fmt.Sprintf("(%s[%s])", ie.Left.String(), ie.Index.String())
}

// HashLiteral represents a hash map literal in the AST.
type HashLiteral struct {
	BaseNode
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) isExpression() {}
func (hl *HashLiteral) String() string {
	pairs := make([]string, 0, len(hl.Pairs))
	for key, value := range hl.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s:%s", key.String(), value.String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}
