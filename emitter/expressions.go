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
			e.write("(")
		}
		e.emit(expr.Left)
		if left < precedence {
			e.write(")")
		}
	}

	e.write(" ")
	e.write(expr.Operator.Text())
	e.write(" ")

	if expr.Right != nil {
		right := Precedence(expr.Right)
		if right < precedence {
			e.write("(")
		}
		e.emit(expr.Right)
		if right < precedence {
			e.write(")")
		}
	}
}

func (e *Emitter) emitCallExpression(expr checker.CallExpression) {
	e.emit(expr.Callee)
	e.write("(")
	defer e.write(")")

	args := expr.Args // This should be ensured by checker
	for i, el := range args.Elements {
		e.emit(el)
		if i != len(args.Elements)-1 {
			e.write(", ")
		}
	}
}

func (e *Emitter) emitComputedAccessExpression(expr checker.ComputedAccessExpression) {
	e.emit(expr.Expr)
	t := expr.Expr.Type()
	if _, ok := t.(checker.List); ok {
		e.write("[")
		e.emit(expr.Property)
		e.write("]")
	}
}

func (e *Emitter) emitFatArrowFunction(f checker.FatArrowFunction) {
	e.emitParams(f.Params)
	e.write(" => ")
	e.emit(f.Body)
}

func (e *Emitter) emitIdentifier(i checker.Identifier) {
	text := i.Token.Text()
	if text == e.thisName {
		e.write("this")
		return
	}
	e.write(getSanitizedName(text))
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

func (e *Emitter) emitListInstance(constructor checker.ListTypeExpression, args []checker.Expression) {
	e.write("[")
	c := constructor.Expr
	max := len(args) - 1
	for i, arg := range args {
		e.emitInstance(c, []checker.Expression{arg})
		if i != max {
			e.write(", ")
		}
	}
	e.write("]")
}
func (e *Emitter) emitObjectInstance(constructor checker.Identifier, args []checker.Expression) {
	e.write("new ")
	e.emit(constructor)
	e.write("(")
	defer e.write(")")
	typing := constructor.Type().(checker.Type).Value.(checker.TypeAlias).Ref.(checker.Object)
	max := len(args) - 1
	i := 0
	members := make([]checker.ObjectExpressionMember, len(args))
	for i := range args {
		members[i] = args[i].(checker.ObjectExpressionMember)
	}
	for name := range typing.Members {
		e.emit(findMemberByName(members, name))
		if i != max {
			e.write(", ")
		}
		i++
	}
}
func (e *Emitter) emitSumInstance(constructor checker.PropertyAccessExpression, args []checker.Expression) {
	e.write("new ")
	e.emit(constructor.Expr)
	e.write("(\"")
	e.emit(constructor.Property)
	e.write("\", ")
	c := e.constructors[constructor.Expr.(checker.Identifier).Text()][constructor.Property.Text()]
	e.emitInstance(c, args)
	e.write(")")
}
func (e *Emitter) emitInstance(constructor checker.Expression, args []checker.Expression) {
	switch c := constructor.(type) {
	case checker.ListTypeExpression:
		e.emitListInstance(c, args)
	case checker.PropertyAccessExpression:
		e.emitSumInstance(c, args)
	case checker.Identifier:
		e.emitObjectInstance(c, args)
	}
}
func (e *Emitter) emitInstanceExpression(expr checker.InstanceExpression) {
	e.emitInstance(expr.Constructor, expr.Args)
}

func (e *Emitter) emitParams(params checker.Params) {
	e.write("(")
	length := len(params.Params)
	for i, param := range params.Params {
		e.emit(param.Identifier)
		if i != length-1 {
			e.write(", ")
		}
	}
	e.write(")")
}

func (e *Emitter) emitPropertyAccessExpression(p checker.PropertyAccessExpression) {
	e.emit(p.Expr)
	if p.Expr.Type().Kind() == checker.TUPLE {
		e.write("[")
		e.emit(p.Property)
		e.write("]")
	} else {
		e.write(".")
		e.emit(p.Property)
	}
}

func (e *Emitter) emitRangeExpression(r checker.RangeExpression) {
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
		if r.Operator.Kind() == tokenizer.RANGE_INCLUSIVE {
			e.write(" + 1")
		}
	} else {
		e.write("1")
	}

	e.write(")")
}

func (e *Emitter) emitSlimArrowFunction(f checker.SlimArrowFunction) {
	e.emitParams(f.Params)
	e.write(" => ")
	e.emit(f.Expr)
}

func (e *Emitter) emitTupleExpression(t checker.TupleExpression) {
	if len(t.Elements) == 1 {
		e.emit(t.Elements[0])
		return
	}
	e.write("[")
	length := len(t.Elements)
	for i, el := range t.Elements {
		e.emit(el)
		if i != length-1 {
			e.write(", ")
		}
	}
	e.write("]")
}
