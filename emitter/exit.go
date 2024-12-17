package emitter

import "github.com/bmelicque/test-parser/parser"

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
		e.emitExpression(r.Value)
	}
	e.write(";\n")
}
