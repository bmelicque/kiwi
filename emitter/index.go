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
	SumFlag
)

type Emitter struct {
	depth        int
	flags        EmitterFlag
	builder      strings.Builder
	thisName     string
	constructors map[string]map[string]parser.Expression
	blockHoister
}

func makeEmitter() *Emitter {
	return &Emitter{
		depth:        0,
		flags:        NoFlags,
		builder:      strings.Builder{},
		constructors: map[string]map[string]parser.Expression{},
		blockHoister: blockHoister{[]hoistedBlock{}},
	}
}

func (e *Emitter) addFlag(flag EmitterFlag) {
	e.flags |= flag
}

func (e *Emitter) hasFlag(flag EmitterFlag) bool {
	return (e.flags & flag) != NoFlags
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

// FIXME: emit vs. emitExpression
func (e *Emitter) emit(node parser.Node) {
	if blocks := e.blockHoister.findStatementBlocks(&node); len(blocks) > 0 {
		for _, block := range blocks {
			e.write(fmt.Sprintf("let %v;\n", block.label))
			e.indent()
			e.emitBlock(block.block)
			e.indent()
		}
	}
	switch node := node.(type) {
	// Statements
	case *parser.Assignment:
		e.emitAssignment(node)
	case *parser.Block:
		label, ok := e.findBlockLabel(node)
		if !ok {
			e.emitBlockExpression(node)
		} else {
			e.write(label)
		}
	case *parser.ForExpression:
		e.emitFor(node)
	case *parser.IfExpression:
		e.emitIfStatement(node)
	case *parser.MatchExpression:
		e.emitMatchStatement(*node)
	case *parser.Exit:
		e.emitExit(node)
	case parser.Expression:
		e.emitExpression(node)
	default:
		panic(fmt.Sprintf("Cannot emit type '%v' (not implemented yet)", reflect.TypeOf(node)))
	}
}

func (e *Emitter) emitExpression(expr parser.Expression) {
	switch expr := expr.(type) {
	case *parser.BinaryExpression:
		e.emitBinaryExpression(expr)
	case *parser.CallExpression:
		e.emitCallExpression(expr)
	case *parser.ComputedAccessExpression:
		e.emitComputedAccessExpression(expr)
	case *parser.FunctionExpression:
		e.emitFunctionExpression(expr)
	case *parser.Identifier:
		e.emitIdentifier(expr)
	case *parser.IfExpression:
		e.emitIfExpression(expr)
	case *parser.Literal:
		e.write(expr.Token.Text())
	case *parser.ParenthesizedExpression:
		e.write("(")
		e.emit(expr.Expr)
		e.write(")")
	case *parser.PropertyAccessExpression:
		e.emitPropertyAccessExpression(expr)
	case *parser.RangeExpression:
		e.emitRangeExpression(expr)
	case *parser.TupleExpression:
		e.emitTupleExpression(expr)
	}
}

func EmitProgram(nodes []parser.Node) string {
	e := makeEmitter()
	for _, node := range nodes {
		findHoisted(node, &e.blockHoister.blocks)
	}

	for _, node := range nodes {
		e.emit(node)
	}
	if e.hasFlag(RangeFlag) {
		e.write("function* _range(start, end) {\n    while (start < end) yield start++;\n}\n")
	}
	if e.hasFlag(SumFlag) {
		e.write("class _Sum {\n")
		e.write("    constructor(_tag, _value) {\n")
		e.write("        this._tag = _tag;\n")
		e.write("        if (arguments.length > 1) { this._value = _value }\n")
		e.write("    }\n}\n")
	}
	return e.string()
}
