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
	SumFlag
)

type Emitter struct {
	depth        int
	flags        EmitterFlag
	builder      strings.Builder
	thisName     string
	constructors map[string]map[string]checker.Expression
	blockHoister
}

func makeEmitter() *Emitter {
	return &Emitter{
		depth:        0,
		flags:        NoFlags,
		builder:      strings.Builder{},
		constructors: map[string]map[string]checker.Expression{},
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

func (e *Emitter) emit(node checker.Node) {
	if blocks := e.blockHoister.findStatementBlocks(&node); len(blocks) > 0 {
		for _, block := range blocks {
			e.write(fmt.Sprintf("let %v;\n", block.label))
			e.indent()
			e.emitBlock(*block.block)
			e.indent()
		}
	}
	switch node := node.(type) {
	// Statements
	case checker.Assignment:
		e.emitAssignment(node)
	case checker.Block:
		label, ok := e.findBlockLabel(&node)
		if !ok {
			e.emitBlockExpression(node)
		} else {
			e.write(label)
		}
	case checker.ExpressionStatement:
		e.emitExpressionStatement(node)
	case checker.For:
		e.emitFor(node)
	case checker.ForRange:
		e.emitForRange(node)
	case checker.If:
		e.emitIfExpression(node)
	case checker.MatchExpression:
		e.emitMatchStatement(node)
	case checker.MethodDeclaration:
		e.emitMethodDeclaration(node)
	case checker.Exit:
		e.emitExit(node)
	case checker.VariableDeclaration:
		e.emitVariableDeclaration(node)

	// Expressions
	case checker.BinaryExpression:
		e.emitBinaryExpression(node)
	case checker.CallExpression:
		e.emitCallExpression(node)
	case checker.ComputedAccessExpression:
		e.emitComputedAccessExpression(node)
	case checker.FunctionExpression:
		e.emitFatArrowFunction(node)
	case checker.Identifier:
		e.emitIdentifier(node)
	case checker.Literal:
		e.write(node.Token.Text())
	case checker.ParenthesizedExpression:
		e.write("(")
		e.emit(node.Expr)
		e.write(")")
	case checker.PropertyAccessExpression:
		e.emitPropertyAccessExpression(node)
	case checker.RangeExpression:
		e.emitRangeExpression(node)
	case checker.TupleExpression:
		e.emitTupleExpression(node)

	default:
		panic(fmt.Sprintf("Cannot emit type '%v' (not implemented yet)", reflect.TypeOf(node)))
	}
}

func EmitProgram(nodes []checker.Node) string {
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
