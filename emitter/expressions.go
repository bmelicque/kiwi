package emitter

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func (e *Emitter) EmitBinaryExpression(expr parser.BinaryExpression) {
	// TODO: function to get JS operator precedence for limiting parenthesis output
	// TODO: if e.minify, replace "===" by "==" (also, make sure that equality check is strict on types)
	e.Write("(")
	if expr.Left != nil {
		e.Emit(expr.Left)
	}

	e.Write(" ")
	e.Write(expr.Operator.Text())
	e.Write(" ")

	if expr.Right != nil {
		e.Emit(expr.Right)
	}
	e.Write(")")
}

func (e *Emitter) EmitFunctionExpression(f parser.FunctionExpression) {
	e.Write("(")
	length := len(f.Params.Elements)
	for i, param := range f.Params.Elements {
		e.Emit(param)
		if i != length-1 {
			e.Write(", ")
		}
	}
	e.Write(")")

	e.Write(" => ")

	if f.Operator.Kind() == tokenizer.SLIM_ARR {
		e.Emit(f.Expr)
	} else { // FAT_ARR
		e.Emit(f.Body)
	}
}

func (e *Emitter) EmitListExpression(l parser.ListExpression) {
	e.Write("[")
	for i, el := range l.Elements {
		e.Emit(el)
		if i != len(l.Elements)-1 {
			e.Write(", ")
		}
	}
	e.Write("]")
}

func (e *Emitter) EmitRangeExpression(r parser.RangeExpression) {
	e.AddFlag(RangeFlag)

	e.Write("range(")

	if r.Left != nil {
		e.Emit(r.Left)
	} else {
		e.Write("0")
	}

	e.Write(", ")

	if r.Right != nil {
		e.Emit(r.Right)
		if r.Operator.Kind() == tokenizer.RANGE_INCLUSIVE {
			e.Write(" + 1")
		}
	} else {
		e.Write("1")
	}

	e.Write(")")
}

func (e *Emitter) EmitTupleExpression(t parser.TupleExpression) {
	e.Write("[")
	length := len(t.Elements)
	for i, el := range t.Elements {
		e.Emit(el)
		if i != length-1 {
			e.Write(", ")
		}
	}
	e.Write("]")
}
