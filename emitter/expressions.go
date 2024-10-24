package emitter

import (
	"github.com/bmelicque/test-parser/parser"
)

func (e *Emitter) emitBinaryExpression(expr *parser.BinaryExpression) {
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

func (e *Emitter) emitBlockExpression(b *parser.Block) {
	label, ok := e.findBlockLabel(b)
	if ok {
		e.write(label)
		return
	}

	if len(b.Statements) == 0 {
		e.write("undefined")
		return
	}
	if len(b.Statements) == 1 {
		e.emit(b.Statements[0])
		return
	}
	e.write("(\n")
	e.depth += 1
	for _, statement := range b.Statements {
		e.indent()
		e.emit(statement)
		e.write(",\n")
	}
	e.depth -= 1
	e.indent()
	e.write(")")
}

func (e *Emitter) emitIfExpression(i *parser.IfExpression) {
	e.emit(i.Condition)
	e.write(" ? ")
	e.emitBlockExpression(i.Body)
	e.write(" : ")
	if i.Alternate == nil {
		e.write("undefined")
		return
	}
	switch alternate := i.Alternate.(type) {
	case *parser.Block:
		e.emitBlockExpression(alternate)
	case *parser.IfExpression:
		e.emitIfExpression(alternate)
	}
}

func findMemberByName(members *parser.TupleExpression, name string) parser.Node {
	for _, member := range members.Elements {
		param := member.(*parser.Param)
		text := param.Identifier.Text()
		if text == name {
			return param.Complement
		}
	}
	return nil
}
func (e *Emitter) emitListInstance(constructor *parser.ListTypeExpression, args *parser.TupleExpression) {
	e.write("[")
	c := constructor.Expr
	max := len(args.Elements) - 1
	for _, arg := range args.Elements[:max] {
		e.emitInstance(
			c,
			&parser.TupleExpression{Elements: []parser.Expression{arg}},
		)
		e.write(", ")
	}
	e.emitInstance(
		c,
		&parser.TupleExpression{Elements: []parser.Expression{args.Elements[max]}},
	)
	e.write("]")
}
func (e *Emitter) emitObjectInstance(constructor *parser.Identifier, args *parser.TupleExpression) {
	e.write("new ")
	e.emit(constructor)
	e.write("(")
	defer e.write(")")
	typing := constructor.Type().(parser.Type).Value.(parser.TypeAlias).Ref.(parser.Object)
	max := len(args.Elements) - 1
	i := 0
	for name := range typing.Members {
		e.emit(findMemberByName(args, name))
		if i != max {
			e.write(", ")
		}
		i++
	}
}
func (e *Emitter) emitSumInstance(constructor *parser.PropertyAccessExpression, args *parser.TupleExpression) {
	e.write("new ")
	e.emit(constructor.Expr)
	e.write("(\"")
	e.emit(constructor.Property)
	e.write("\", ")
	sum := constructor.Expr.(*parser.Identifier).Text()
	cons := constructor.Property.(*parser.Identifier).Text()
	c := e.constructors[sum][cons]
	e.emitInstance(c, args)
	e.write(")")
}
func (e *Emitter) emitInstance(constructor parser.Expression, args *parser.TupleExpression) {
	switch c := constructor.(type) {
	case *parser.ListTypeExpression:
		e.emitListInstance(c, args)
	case *parser.PropertyAccessExpression:
		e.emitSumInstance(c, args)
	case *parser.Identifier:
		e.emitObjectInstance(c, args)
	}
}
func (e *Emitter) emitInstanceExpression(expr *parser.CallExpression) {
	e.emitInstance(expr.Callee, expr.Args.Expr.(*parser.TupleExpression))
}
func (e *Emitter) emitCallExpression(expr *parser.CallExpression) {
	if expr.Callee.Type().Kind() == parser.TYPE {
		e.emitInstanceExpression(expr)
		return
	}

	e.emit(expr.Callee)
	e.write("(")
	defer e.write(")")

	args := expr.Args.Expr.(*parser.TupleExpression).Elements
	max := len(args) - 1
	for i := range args[:max] {
		e.emit(args[i])
		e.write(", ")
	}
	e.emit(args[max])
}

func (e *Emitter) emitComputedAccessExpression(expr *parser.ComputedAccessExpression) {
	e.emit(expr.Expr)
	t := expr.Expr.Type()
	if _, ok := t.(parser.List); ok {
		e.write("[")
		e.emit(expr.Property.Expr)
		e.write("]")
	}
}

func (e *Emitter) emitFunctionExpression(f *parser.FunctionExpression) {
	e.write("(")
	args := f.Params.Expr.(*parser.TupleExpression).Elements
	max := len(args)
	for i := range args[:max] {
		param := args[i].(*parser.Param)
		e.emit(param.Identifier)
		e.write(", ")
	}
	e.emit(args[max].(*parser.Param).Identifier)
	e.write(") => ")
	e.emit(f.Body)
}

func (e *Emitter) emitIdentifier(i *parser.Identifier) {
	text := i.Token.Text()
	if text == e.thisName {
		e.write("this")
		return
	}
	e.write(getSanitizedName(text))
}

func (e *Emitter) emitPropertyAccessExpression(p *parser.PropertyAccessExpression) {
	e.emit(p.Expr)
	if p.Expr.Type().Kind() == parser.TUPLE {
		e.write("[")
		e.emit(p.Property)
		e.write("]")
	} else {
		e.write(".")
		e.emit(p.Property)
	}
}

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

func (e *Emitter) emitTryExpression(t *parser.TryExpression) {
	e.emitExpression(t.Expr)
}

func (e *Emitter) emitTupleExpression(t *parser.TupleExpression) {
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

func (e *Emitter) emitUnaryExpression(u *parser.UnaryExpression) {
	if u.Operator.Kind() == parser.Bang && u.Type().Kind() == parser.BOOLEAN {
		e.write("!")
		e.emitExpression(u.Operand)
	}
}
