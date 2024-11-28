package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

const maxClassParamsLength = 66

func (e *Emitter) emitFor(f *parser.ForExpression) {
	a, ok := f.Statement.(*parser.Assignment)
	if !ok {
		e.write("while (")
		e.emit(f.Statement)
		e.write(") ")
		e.emitBlockStatement(f.Body)
	}

	e.write("for (let ")
	// FIXME: tuples...
	e.emit(a.Pattern)
	e.write(" of ")
	e.emit(a.Value)
	e.write(") ")
	e.emitBlockStatement(f.Body)
}

func (e *Emitter) emitMatchStatement(m parser.MatchExpression) {
	// TODO: break outer loop
	// TODO: declare _m only if calling something
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

func (e *Emitter) emitExit(r *parser.Exit) {
	switch r.Operator.Kind() {
	case parser.BreakKeyword:
		e.write("break")
	case parser.ContinueKeyword:
		e.write("continue")
	case parser.ReturnKeyword:
		e.write("return")
	case parser.ThrowKeyword:
		e.write("throw")
	}
	if r.Value != nil {
		e.write(" ")
		e.emit(r.Value)
	}
	e.write(";\n")
}
