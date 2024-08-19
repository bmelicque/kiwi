package emitter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/bmelicque/test-parser/checker"
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

func (e *Emitter) Emit(node interface{}) {
	switch node := node.(type) {
	// Statements
	case checker.Assignment:
		e.emitAssignment(node)
	case checker.Body:
		e.emitBody(node)
	case checker.ExpressionStatement:
		e.emitExpressionStatement(node)
	case checker.For:
		e.emitFor(node)
	case checker.ForRange:
		e.emitForRange(node)
	case checker.If:
		e.emitIf(node)
	case checker.MethodDeclaration:
		e.emitMethodDeclaration(node)
	case checker.Return:
		e.emitReturn(node)
	case checker.VariableDeclaration:
		e.emitVariableDeclaration(node)

	// Expressions
	case checker.BinaryExpression:
		e.emitBinaryExpression(node)
	case checker.CallExpression:
		e.emitCallExpression(node)
	case checker.FatArrowFunction:
		e.emitFatArrowFunction(node)
	case checker.Identifier:
		text := node.Token.Text()
		if text == e.thisName {
			e.Write("this")
		} else {
			e.Write(text)
		}
	case checker.ListExpression:
		e.emitListExpression(node)
	case checker.Literal:
		e.Write(node.Token.Text())
	case checker.ObjectExpression:
		e.emitObjectExpression(node)
	case checker.PropertyAccessExpression:
		e.emitPropertyAccessExpression(node)
	case checker.RangeExpression:
		e.emitRangeExpression(node)
	case checker.SlimArrowFunction:
		e.emitSlimArrowFunction(node)
	case checker.TupleExpression:
		e.emitTupleExpression(node)

	default:
		panic(fmt.Sprintf("Cannot emit type '%v' (not implemented yet)", reflect.TypeOf(node)))
	}
}
