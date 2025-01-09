package emitter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/bmelicque/test-parser/parser"
)

type EmitterFlag int

const (
	NoFlags EmitterFlag = 0
	SumFlag EmitterFlag = 1 << iota
	RefComparisonFlag
	ObjectComparisonFlag

	AllFlags
)

type Emitter struct {
	depth        int
	flags        EmitterFlag
	builder      strings.Builder
	thisName     string
	constructors map[string]map[string]parser.Expression
	uninlinables map[parser.Node]int
}

func makeEmitter() *Emitter {
	return &Emitter{
		depth:        0,
		flags:        NoFlags,
		builder:      strings.Builder{},
		constructors: map[string]map[string]parser.Expression{},
		uninlinables: map[parser.Node]int{},
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

func (e *Emitter) emit(node parser.Node) {
	//TODO: if not node that needs extraction, look if contains one
	if !isUninlinable(node) {
		e.extractUninlinables(node)
	}
	switch node := node.(type) {
	// Statements
	case *parser.Assignment:
		e.emitAssignment(node)
	case *parser.Block:
		e.emitBlockStatement(node)
	case *parser.CatchExpression:
		e.emitCatchStatement(node)
	case *parser.ForExpression:
		e.emitFor(node)
	case *parser.IfExpression:
		e.emitIfStatement(node)
	case *parser.MatchExpression:
		e.emitMatchStatement(*node)
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
		e.write(fmt.Sprintf("_tmp%v", id))
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

func EmitProgram(nodes []parser.Node) string {
	e := makeEmitter()
	e.scan(nodes)
	if e.hasFlag(SumFlag) {
		e.write("class _Sum {\n")
		e.write("    constructor(_tag, _value) {\n")
		e.write("        this._tag = _tag;\n")
		e.write("        if (arguments.length > 1) { this._value = _value }\n")
		e.write("    }\n}\n")
	}
	if e.hasFlag(RefComparisonFlag) {
		e.write("function __refEquals(a, b) { return a(4) == b(4) && a(2) == b(2) }\n")
	}
	if e.hasFlag(ObjectComparisonFlag) {
		e.write("var __equals=(a,b)=>typeof a==typeof b&&(typeof a!=\"object\"||a==null||b==null?a==b:a.constructor==b.constructor&&(!Array.isArray(a)||a.length==b.length)&&!Object.keys(a).find(k=>!__equals(a[k],b[k]))));\n")
	}

	for _, node := range nodes {
		e.emit(node)
	}

	return e.string()
}

func (e *Emitter) scan(nodes []parser.Node) {
	stop := false
	addFlag := func(flag EmitterFlag) {
		e.addFlag(flag)
		if e.hasFlag(AllFlags - 1) {
			stop = true
		}
	}

	handleNode := func(n parser.Node, skip func()) {
		if stop {
			skip()
		}
		switch {
		case isSumDef(n):
			addFlag(SumFlag)
		case isRefComparison(n):
			addFlag(RefComparisonFlag)
		case isObjectComparison(n):
			addFlag(ObjectComparisonFlag)
		}
	}

	for _, node := range nodes {
		parser.Walk(node, handleNode)
		if stop {
			break
		}
	}
}

func isSumDef(n parser.Node) bool {
	a, ok := n.(*parser.Assignment)
	if !ok {
		return false
	}
	if a.Operator.Kind() != parser.Define {
		return false
	}
	t, ok := a.Value.Type().(parser.Type)
	if !ok {
		return false
	}
	_, isSum := t.Value.(parser.Sum)
	return isSum
}

func isRefComparison(n parser.Node) bool {
	b, ok := n.(*parser.BinaryExpression)
	if !ok {
		return false
	}
	if b.Operator.Kind() != parser.Equal {
		return false
	}
	_, ok = b.Left.Type().(parser.Ref)
	return ok
}

func isObjectComparison(n parser.Node) bool {
	b, ok := n.(*parser.BinaryExpression)
	if !ok {
		return false
	}
	if b.Operator.Kind() != parser.Equal {
		return false
	}
	switch b.Left.Type().(type) {
	case parser.List, parser.Trait, parser.TypeAlias:
		return true
	default:
		return false
	}
}
