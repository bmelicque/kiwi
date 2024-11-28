package emitter

import "github.com/bmelicque/test-parser/parser"

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
