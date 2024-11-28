package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitRangeExpression(r *parser.RangeExpression) {
	e.addFlag(RangeFlag)

	e.write("_range(")

	if r.Left != nil {
		e.emit(r.Left)
	} else {
		e.write("0")
	}

	e.write(", ")

	if r.Right != nil {
		e.emit(r.Right)
		if r.Operator.Kind() == parser.InclusiveRange {
			e.write(" + 1")
		}
	} else {
		e.write("1")
	}

	e.write(")")
}
