package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

func (e *Emitter) emitMatchStatement(m parser.MatchExpression) {
	e.write("const _m = ")
	e.emit(m.Value)
	e.write(";\n")
	if _, ok := m.Value.Type().(parser.Sum); ok {
		e.write("switch (_m._tag) {\n")
	} else {
		e.write("switch (_m.constructor) {\n")
	}
	for _, c := range m.Cases {
		e.indent()
		if c.IsCatchall() {
			e.write("default:")
		} else if call, ok := c.Pattern.(*parser.CallExpression); ok {
			e.write("case ")
			e.emit(call.Callee)
			e.write(": {\n")
		} else if id, ok := c.Pattern.(*parser.Identifier); ok {
			e.write("case ")
			e.emit(id)
			e.write(": {\n")
		}
		e.depth++
		if c.Pattern != nil {
			id := c.Pattern.(*parser.Identifier)
			e.indent()
			if _, ok := m.Value.Type().(parser.Sum); ok {
				e.write(fmt.Sprintf("let %v = _m._value;\n", id.Text()))
			} else {
				e.write(fmt.Sprintf("let %v = _m;\n", id.Text()))
			}
		}
		for _, s := range c.Statements {
			e.indent()
			e.emit(s)
		}
		e.indent()
		e.write("break;\n")
		e.indent()
		e.write("}\n")
		e.depth--
	}
	e.indent()
	e.write("}\n")
}
