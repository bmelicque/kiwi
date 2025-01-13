package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

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

func findMemberByName(members *parser.TupleExpression, name string) parser.Expression {
	for _, member := range members.Elements {
		switch member := member.(type) {
		case *parser.Param:
			text := member.Identifier.Text()
			if text == name {
				return member.Complement
			}
		case *parser.Entry:
			text := member.Key.(*parser.Identifier).Text()
			if text == name {
				return member.Value
			}
		}
	}
	return nil
}
func (e *Emitter) emitObjectInstance(constructor *parser.Identifier, args *parser.TupleExpression) {
	if len(args.Elements) == 0 {
		e.write("new ")
		e.emitExpression(constructor)
		e.write("()")
		return
	}

	e.write("new ")
	e.emitExpression(constructor)
	e.write("(")
	typing := constructor.Type().(parser.Type).Value.(parser.TypeAlias).Ref.(parser.Object)
	l := len(args.Elements)
	i := 0
	for _, m := range append(typing.Members, typing.Defaults...) {
		member := findMemberByName(args, m.Name)
		if member == nil {
			e.write("undefined")
		} else {
			e.emitExpression(member)
			i++
		}
		if i != l {
			e.write(", ")
		} else {
			break
		}
	}
	e.write(")")
}
func (e *Emitter) emitSumInstance(constructor *parser.PropertyAccessExpression, args *parser.TupleExpression) {
	e.write("new ")
	e.emitExpression(constructor.Expr)
	e.write("(\"")
	e.emitExpression(constructor.Property)
	e.write("\", ")
	sum := constructor.Expr.(*parser.Identifier).Text()
	cons := constructor.Property.(*parser.Identifier).Text()
	c := e.constructors[sum][cons]
	e.emitInstance(c, args)
	e.write(")")
}

func (e *Emitter) emitRefInstance(constructor parser.Expression, args *parser.TupleExpression) {
	c := constructor.Type().(parser.Type).Value.(parser.Ref).To
	fmt.Println(c.Text())
	if implementsNode(c) {
		e.write("new __.NodePointer(")
	} else {
		e.write("new __.Pointer(null, ")
	}
	e.emitInstance(constructor.(*parser.UnaryExpression).Operand, args)
	e.write(")")
}

func (e *Emitter) emitInstance(constructor parser.Expression, args *parser.TupleExpression) {
	if isReferenceExpression(constructor) {
		e.emitRefInstance(constructor, args)
		return
	}
	switch c := constructor.(type) {
	case *parser.ListTypeExpression:
		e.emitListInstance(c, args)
	case *parser.PropertyAccessExpression:
		e.emitSumInstance(c, args)
	case *parser.ComputedAccessExpression:
		e.emitInstance(c.Expr, args)
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
