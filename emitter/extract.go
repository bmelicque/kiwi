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
		if isUninlinable(node) {
			e.uninlinables[node] = len(e.uninlinables)
			skip()
		}
	})
}

func isUninlinable(node parser.Node) bool {
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
		e.write(fmt.Sprintf("let _tmp%v;\n", id))
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
		e.write(";\n")
	}
	e.indent()
	e.write(fmt.Sprintf("_tmp%v = ", id))
	e.emit(b.Statements[max])
	e.write(";\n")
	e.depth--
	e.indent()
	e.write("}\n")
}

func emitExtractedCatch(e *Emitter, c *parser.CatchExpression) {
	e.write("try {\n")
	e.depth++
	e.indent()
	id := e.uninlinables[c]
	e.write(fmt.Sprintf("_tmp%v = ", id))
	e.emit(c.Left)
	e.write(";\n")
	e.depth--
	e.write("} catch (")
	e.emit(c.Identifier)
	e.write(") ")
	emitExtractedBlock(e, c.Body, id)
}
