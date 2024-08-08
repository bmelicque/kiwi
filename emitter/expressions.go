package emitter

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func (e *Emitter) EmitBinaryExpression(expr parser.BinaryExpression) {
	precedence := Precedence(expr)
	if expr.Left != nil {
		left := Precedence(expr.Left)
		if left < precedence {
			e.Write("(")
		}
		e.Emit(expr.Left)
		if left < precedence {
			e.Write(")")
		}
	}

	e.Write(" ")
	e.Write(expr.Operator.Text())
	e.Write(" ")

	if expr.Right != nil {
		right := Precedence(expr.Right)
		if right < precedence {
			e.Write("(")
		}
		e.Emit(expr.Right)
		if right < precedence {
			e.Write(")")
		}
	}
}

func (e *Emitter) EmitCallExpression(expr parser.CallExpression) {
	e.Emit(expr.Callee)
	e.Write("(")
	defer e.Write(")")

	args := expr.Args.(parser.TupleExpression) // This should be ensured by checker
	for i, el := range args.Elements {
		e.Emit(el)
		if i != len(args.Elements)-1 {
			e.Write(", ")
		}
	}
}

func (e *Emitter) EmitFunctionExpression(f parser.FunctionExpression) {
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
		e.Emit(*f.Body)
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

func findMemberByName(members []parser.Node, name string) parser.Node {
	for _, member := range members {
		expr := member.(parser.TypedExpression)
		text := expr.Expr.(*parser.TokenExpression).Token.Text()
		if text == name {
			return expr.Typing
		}
	}
	return nil
}

func (e *Emitter) EmitObjectExpression(o parser.ObjectExpression) {
	e.Emit(o.Typing)
	e.Write("(")
	defer e.Write(")")
	typing := o.Typing.Type().(parser.TypeRef).Ref.(parser.Object)
	max := len(o.Members) - 1
	i := 0
	for name := range typing.Members {
		e.Emit(findMemberByName(o.Members, name))
		if i != max {
			e.Write(", ")
		}
		i++
	}
}

func (e *Emitter) EmitPropertyAccessExpression(p *parser.PropertyAccessExpression) {
	e.Emit(p.Expr)
	e.Write(".")
	e.Emit(p.Property)
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
	if len(t.Elements) == 1 {
		e.Emit(t.Elements[0])
		return
	}
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
