package emitter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/bmelicque/test-parser/parser"
)

type EmitterFlag int

const (
	NoFlags   EmitterFlag = 0
	RangeFlag EmitterFlag = 1 << iota
)

type Emitter struct {
	depth    int
	flags    EmitterFlag
	builder  strings.Builder
	thisName string
}

func MakeEmitter() *Emitter {
	return &Emitter{
		depth:   0,
		flags:   NoFlags,
		builder: strings.Builder{},
	}
}

func (e *Emitter) AddFlag(flag EmitterFlag) {
	e.flags |= flag
}

func (e *Emitter) Write(str string) {
	e.builder.WriteString(str)
}

func (e *Emitter) Indent() {
	for i := 0; i < e.depth; i++ {
		e.builder.WriteString("    ")
	}
}

func (e Emitter) String() string {
	return e.builder.String()
}

func (e *Emitter) Emit(node parser.Node) {
	switch node := node.(type) {
	// Statements
	case parser.Assignment:
		e.EmitAssignment(node)
	case parser.Body:
		e.EmitBody(node)
	case parser.ExpressionStatement:
		e.EmitExpressionStatement(node)
	case parser.For:
		e.EmitFor(node)
	case parser.IfElse:
		e.EmitIfElse(node)
	case parser.Return:
		e.EmitReturn(node)

	// Expressions
	case parser.BinaryExpression:
		e.EmitBinaryExpression(node)
	case parser.CallExpression:
		e.EmitCallExpression(node)
	case parser.FunctionExpression:
		e.Write("(")
		e.EmitFunctionExpression(node)
	case parser.ListExpression:
		e.EmitListExpression(node)
	case parser.ObjectExpression:
		e.EmitObjectExpression(node)
	case *parser.PropertyAccessExpression:
		e.EmitPropertyAccessExpression(node)
	case parser.RangeExpression:
		e.EmitRangeExpression(node)
	case *parser.TokenExpression:
		text := node.Token.Text()
		if text == e.thisName {
			e.Write("this")
		} else {
			e.Write(text)
		}
	case parser.TupleExpression:
		e.EmitTupleExpression(node)
	case parser.TypedExpression:
		e.Emit(node.Expr)

	default:
		panic(fmt.Sprintf("Cannot emit type '%v' (not implemented yet)", reflect.TypeOf(node)))
	}
}
