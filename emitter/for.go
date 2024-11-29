package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitFor(f *parser.ForExpression) {
	if f.Expr == nil {
		e.write("while (true) ")
		e.emitBlockStatement(f.Body)
		return
	}

	binary, ok := f.Expr.(*parser.BinaryExpression)
	if !ok || binary.Operator.Kind() != parser.InKeyword {
		e.write("while (")
		e.emit(f.Expr)
		e.write(") ")
		e.emitBlockStatement(f.Body)
		return
	}
	if _, ok := binary.Right.(*parser.RangeExpression); ok {
		if _, ok := binary.Left.(*parser.TupleExpression); ok {
			emitForRangeTuple(e, f)
		} else {
			emitForRange(e, f)
		}
		return
	}
	e.write("for (let ")
	// FIXME: tuples...
	e.emit(binary.Left)
	e.write(" of ")
	e.emit(binary.Right)
	e.write(") ")
	e.emitBlockStatement(f.Body)
}

func emitForRange(e *Emitter, f *parser.ForExpression) {
	binary := f.Expr.(*parser.BinaryExpression)
	r := binary.Right.(*parser.RangeExpression)
	identifier := binary.Left.(*parser.Identifier)

	e.write("for (let ")
	e.emitIdentifier(identifier)
	e.write(" = ")
	e.emitExpression(r.Left)
	e.write("; ")
	e.emitExpression(identifier)
	if r.Operator.Kind() == parser.InclusiveRange {
		e.write(" <= ")
	} else {
		e.write(" < ")
	}
	e.emitExpression(r.Right)
	e.write("; ")
	e.emitExpression(identifier)
	e.write("++) ")
	e.emitBlockStatement(f.Body)
}

func emitForRangeTuple(e *Emitter, f *parser.ForExpression) {
	binary := f.Expr.(*parser.BinaryExpression)
	r := binary.Right.(*parser.RangeExpression)
	tuple := binary.Left.(*parser.TupleExpression)

	e.write("for (let ")
	e.emitExpression(tuple.Elements[0])
	e.write(" = ")
	e.emitExpression(r.Left)
	e.write(", ")
	e.emitExpression(tuple.Elements[1])
	e.write(" = 0; ")
	e.emitExpression(tuple.Elements[0])
	if r.Operator.Kind() == parser.InclusiveRange {
		e.write(" <= ")
	} else {
		e.write(" < ")
	}
	e.emitExpression(r.Right)
	e.write("; ")
	e.emitExpression(tuple.Elements[0])
	e.write("++, ")
	e.emitExpression(tuple.Elements[1])
	e.write("++) ")
	e.emitBlockStatement(f.Body)
}
