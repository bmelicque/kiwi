package emitter

import (
	"github.com/bmelicque/test-parser/parser"
)

func (e *Emitter) emitTupleExpression(t *parser.TupleExpression) {
	if len(t.Elements) == 1 {
		e.emitExpression(t.Elements[0])
		return
	}
	e.write("[")
	length := len(t.Elements)
	for i, el := range t.Elements {
		e.emitExpression(el)
		if i != length-1 {
			e.write(", ")
		}
	}
	e.write("]")
}
