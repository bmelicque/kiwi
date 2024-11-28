package emitter

import "github.com/bmelicque/test-parser/parser"

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
func (e *Emitter) emitMapInstance(args *parser.TupleExpression) {
	if len(args.Elements) == 0 {
		e.write("new Map()")
		return
	}
	e.write("new Map([")

	max := len(args.Elements) - 1
	for _, arg := range args.Elements[:max] {
		emitMapEntry(e, arg)
		e.write(", ")
	}
	emitMapEntry(e, args.Elements[max])

	e.write("])")
}
func emitMapEntry(e *Emitter, arg parser.Expression) {
	entry := arg.(*parser.Entry)
	var key parser.Expression
	if b, ok := entry.Key.(*parser.BracketedExpression); ok {
		key = b.Expr
	} else {
		key = entry.Key
	}

	e.write("[")
	e.emitExpression(key)
	e.write(", ")
	e.emitExpression(entry.Value)
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
		if c.Text() == "Map" {
			e.emitMapInstance(args)
		} else {
			e.emitObjectInstance(c, args)
		}
	}
}
func (e *Emitter) emitInstanceExpression(expr *parser.InstanceExpression) {
	e.emitInstance(expr.Typing, expr.Args.Expr.(*parser.TupleExpression))
}
