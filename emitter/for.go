package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitFor(f *parser.ForExpression) {
	binary, ok := f.Expr.(*parser.BinaryExpression)
	if ok && binary.Operator.Kind() == parser.InKeyword {
		e.write("for (let ")
		// FIXME: tuples...
		e.emit(binary.Left)
		e.write(" of ")
		e.emit(binary.Right)
		e.write(") ")
		e.emitBlockStatement(f.Body)
	} else {
		e.write("while (")
		e.emit(f.Expr)
		e.write(") ")
		e.emitBlockStatement(f.Body)
	}
}
