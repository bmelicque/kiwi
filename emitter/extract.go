package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

func (e *Emitter) findUninlinables(node parser.Node) {
	parser.Walk(node, func(node parser.Node, skip func()) {
		if _, ok := node.(*parser.FunctionExpression); ok {
			skip()
			return
		}
		if _, ok := node.(*parser.InstanceExpression); ok {
			skip()
			return
		}
		if isTypeDef(node) {
			skip()
			return
		}
		if needsEscape(node) {
			e.uninlinables[node] = len(e.uninlinables)
			skip()
		}
	})
}

func isTypeDef(node parser.Node) bool {
	a, ok := node.(*parser.Assignment)
	return ok && a.Operator.Kind() == parser.Define && isTypePattern(a.Pattern)
}

func needsEscape(node parser.Node) bool {
	switch n := node.(type) {
	case *parser.CatchExpression,
		*parser.ForExpression,
		*parser.IfExpression,
		*parser.MatchExpression:
		return true
	case *parser.Block:
		return len(n.Statements) >= 2
	default:
		return false
	}
}

func (e *Emitter) extractUninlinables(node parser.Node) {
	startAt := len(e.uninlinables)
	e.findUninlinables(node)
	for n, id := range e.uninlinables {
		if id < startAt {
			continue
		}
		// outline block
		e.write(fmt.Sprintf("let __tmp%v;\n", id))
		e.indent()
		switch n := n.(type) {
		case *parser.Block:
			emitExtractedBlock(e, n, id)
		case *parser.CatchExpression:
			emitExtractedCatch(e, n)
		}
		e.indent()
	}
}

func emitExtractedBlock(e *Emitter, b *parser.Block, id int) {
	e.write("{\n")
	e.depth++
	max := len(b.Statements) - 1
	for _, statement := range b.Statements[:max] {
		e.indent()
		e.emit(statement)
	}
	e.indent()
	last := b.Statements[max]
	if !parser.IsExiting(last) {
		e.write(fmt.Sprintf("__tmp%v = ", id))
	}
	e.emit(last)
	e.depth--
	e.indent()
	e.write("}\n")
}

func emitExtractedCatch(e *Emitter, c *parser.CatchExpression) {
	e.write("try {\n")
	e.depth++
	e.indent()
	id := e.uninlinables[c]
	e.write(fmt.Sprintf("__tmp%v = ", id))
	e.emitExpression(c.Left)
	e.write(";\n")
	e.depth--
	e.write("} catch (")
	if c.Identifier != nil {
		e.emitExpression(c.Identifier)
	} else {
		e.write("_")
	}
	e.write(") ")
	emitExtractedBlock(e, c.Body, id)
}
