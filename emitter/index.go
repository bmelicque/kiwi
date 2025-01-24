package emitter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/bmelicque/test-parser/parser"
)

type Emitter struct {
	depth        int
	builder      strings.Builder
	thisName     string
	constructors map[string]map[string]parser.Expression
	uninlinables map[parser.Node]int
	stdEmitter
}

func makeEmitter() *Emitter {
	return &Emitter{
		depth:        0,
		builder:      strings.Builder{},
		constructors: map[string]map[string]parser.Expression{},
		uninlinables: map[parser.Node]int{},
	}
}

func (e *Emitter) write(str string) {
	e.builder.WriteString(str)
}

func (e *Emitter) indent() {
	for i := 0; i < e.depth; i++ {
		e.builder.WriteString("    ")
	}
}

func (e Emitter) string() string {
	return e.builder.String()
}

func (e *Emitter) emitAtTopLevel(node parser.Node) {
	switch node := node.(type) {
	case *parser.Assignment:
		e.extractUninlinables(node)
		e.emitAssignment(node, true)
	default:
		e.emit(node)
	}
}

func (e *Emitter) emit(node parser.Node) {
	//TODO: if not node that needs extraction, look if contains one
	if !isUninlinable(node) {
		e.extractUninlinables(node)
	}
	switch node := node.(type) {
	// Statements
	case *parser.Assignment:
		e.emitAssignment(node, false)
	case *parser.Block:
		e.emitBlockStatement(node)
	case *parser.CatchExpression:
		e.emitCatchStatement(node)
	case *parser.ForExpression:
		e.emitFor(node)
	case *parser.IfExpression:
		e.emitIfStatement(node)
	case *parser.MatchExpression:
		e.emitMatchStatement(node)
	case *parser.Exit:
		e.emitExit(node)
	case *parser.UseDirective:
		e.emitUseStatement(node)
	case parser.Expression:
		e.emitExpression(node)
		e.write(";\n")
	default:
		panic(fmt.Sprintf("Cannot emit type '%v' (not implemented yet)", reflect.TypeOf(node)))
	}
}

func (e *Emitter) emitExpression(expr parser.Expression) {
	switch expr := expr.(type) {
	case *parser.Block:
		e.emitBlockExpression(expr)
	case *parser.BinaryExpression:
		e.emitBinaryExpression(expr)
	case *parser.CallExpression:
		e.emitCallExpression(expr, true)
	case *parser.CatchExpression:
		id, ok := e.uninlinables[expr]
		if !ok {
			panic("Catch expression should have been extracted!")
		}
		e.write(fmt.Sprintf("__tmp%v", id))
		delete(e.uninlinables, expr)
	case *parser.ComputedAccessExpression:
		e.emitComputedAccessExpression(expr)
	case *parser.FunctionExpression:
		e.emitFunctionExpression(expr)
	case *parser.Identifier:
		e.emitIdentifier(expr)
	case *parser.IfExpression:
		e.emitIfExpression(expr)
	case *parser.InstanceExpression:
		e.emitInstanceExpression(expr)
	case *parser.Literal:
		e.write(expr.Token.Text())
	case *parser.ParenthesizedExpression:
		e.write("(")
		e.emit(expr.Expr)
		e.write(")")
	case *parser.PropertyAccessExpression:
		e.emitPropertyAccessExpression(expr, false)
	case *parser.TupleExpression:
		e.emitTupleExpression(expr)
	case *parser.UnaryExpression:
		e.emitUnaryExpression(expr)
	}
}

func EmitProgram(program parser.Program) (string, StandardFlags) {
	e := makeEmitter()
	e.write("const ")
	emitScope(e, program.Scope())
	e.write(" = {};\n")
	for _, node := range program.Nodes() {
		e.emitAtTopLevel(node)
	}
	return e.string(), e.flags
}
