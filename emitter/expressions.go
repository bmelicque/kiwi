package emitter

import (
	"github.com/bmelicque/test-parser/checker"
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func (e *Emitter) emitBinaryExpression(expr checker.BinaryExpression) {
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

func (e *Emitter) emitCallExpression(expr checker.CallExpression) {
	e.Emit(expr.Callee)
	e.Write("(")
	defer e.Write(")")

	args := expr.Args // This should be ensured by checker
	for i, el := range args.Elements {
		e.Emit(el)
		if i != len(args.Elements)-1 {
			e.Write(", ")
		}
	}
}

func (e *Emitter) emitFatArrowFunction(f checker.FatArrowFunction) {
	e.emitParams(f.Params)
	e.Write(" => ")
	e.Emit(f.Body)
}

func (e *Emitter) emitListExpression(l checker.ListExpression) {
	e.Write("[")
	for i, el := range l.Elements {
		e.Emit(el)
		if i != len(l.Elements)-1 {
			e.Write(", ")
		}
	}
	e.Write("]")
}

func findMemberByName(members []checker.ObjectExpressionMember, name string) parser.Node {
	for _, member := range members {
		text := member.Name.Token.Text()
		if text == name {
			return member.Value
		}
	}
	return nil
}

func (e *Emitter) emitObjectExpression(o checker.ObjectExpression) {
	e.Emit(o.Typing)
	e.Write("(")
	defer e.Write(")")
	typing := o.Typing.Type().(checker.TypeRef).Ref.(checker.Object)
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

func (e *Emitter) emitParams(params checker.Params) {
	e.Write("(")
	length := len(params.Params)
	for i, param := range params.Params {
		e.Emit(param.Identifier)
		if i != length-1 {
			e.Write(", ")
		}
	}
	e.Write(")")
}

func (e *Emitter) emitPropertyAccessExpression(p checker.PropertyAccessExpression) {
	e.Emit(p.Expr)
	e.Write(".")
	e.Emit(p.Property)
}

func (e *Emitter) emitRangeExpression(r checker.RangeExpression) {
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

func (e *Emitter) emitSlimArrowFunction(f checker.SlimArrowFunction) {
	e.emitParams(f.Params)
	e.Write(" => ")
	e.Emit(f.Expr)
}

func (e *Emitter) emitTupleExpression(t checker.TupleExpression) {
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
