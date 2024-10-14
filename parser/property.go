package parser

import (
	"fmt"
	"strconv"
)

// Expr.Property
type PropertyAccessExpression struct {
	Expr     Expression
	Property Expression
	typing   ExpressionType
}

func (p *PropertyAccessExpression) Loc() Loc {
	return Loc{
		Start: p.Expr.Loc().Start,
		End:   p.Property.Loc().End,
	}
}
func (p *PropertyAccessExpression) Type() ExpressionType { return p.typing }

func (expr *PropertyAccessExpression) typeCheck(p *Parser) {
	switch expr.Expr.Type().(type) {
	case Tuple:
		typeCheckTupleIndexAccess(p, expr)
	case Type:
		typeCheckSumConstructorAccess(p, expr)
	default:
		typeCheckPropertyAccess(p, expr)
	}
}

func parsePropertyAccess(p *Parser, left Expression) Expression {
	p.Consume() // .
	if _, ok := left.(*ParenthesizedExpression); ok {
		if p.Peek().Kind() == LeftParenthesis {
			parseTraitExpression(p, left)
		}
	}
	prop := fallback(p)
	switch prop.(type) {
	case *Identifier, *Literal:
	default:
		p.report("Key name expected", prop.Loc())
	}
	return &PropertyAccessExpression{Expr: left, Property: prop}
}

// check accessing a tuple's index: tuple.0
func typeCheckTupleIndexAccess(p *Parser, expr *PropertyAccessExpression) {
	property, ok := expr.Property.(*Literal)
	if !ok || property.Type().Kind() != NUMBER {
		expr.typing = Primitive{UNKNOWN}
		return
	}
	number, err := strconv.Atoi(property.Text())
	if err != nil {
		p.report("Integer expected", property.Loc())
		expr.typing = Primitive{UNKNOWN}
		return
	}
	elements := expr.Expr.Type().(Tuple).elements
	if number > len(elements)-1 || number < 0 {
		p.report("Index out of range", property.Loc())
		expr.typing = Primitive{UNKNOWN}
		return
	}
	expr.typing = elements[number]
}

// check accessing a sum type's subconstructor: SumType.Constructor
func typeCheckSumConstructorAccess(p *Parser, expr *PropertyAccessExpression) {
	property, ok := expr.Property.(*Identifier)
	if !ok {
		expr.typing = Primitive{UNKNOWN}
		return
	}
	name := property.Token.Text()

	expr.typing = getSumTypeConstructor(expr.Expr.Type().(Type), name)
	if expr.typing == (Primitive{UNKNOWN}) {
		p.report(
			fmt.Sprintf("Property '%v' doesn't exist on this type", name),
			expr.Property.Loc(),
		)
	}
}

func getSumTypeConstructor(t Type, name string) ExpressionType {
	alias, ok := t.Value.(TypeAlias)
	if !ok {
		return Primitive{UNKNOWN}
	}

	sum, ok := alias.Ref.(Sum)
	if !ok {
		return Primitive{UNKNOWN}
	}

	constructor, ok := sum.Members[name]
	if !ok {
		return Primitive{UNKNOWN}
	}

	if constructor == nil {
		return alias
	}
	return *constructor
}

// check accessing an object's property or method: object.property
func typeCheckPropertyAccess(p *Parser, expr *PropertyAccessExpression) {
	property, ok := expr.Property.(*Identifier)
	if expr.Property != nil && !ok {
		expr.typing = Primitive{UNKNOWN}
		return
	}
	var name string
	if property != nil {
		name = property.Token.Text()
	}

	alias, ok := expr.Expr.Type().(TypeAlias)
	if !ok {
		p.report(
			fmt.Sprintf("Property '%v' doesn't exist on this type", name),
			expr.Property.Loc(),
		)
		expr.typing = Primitive{UNKNOWN}
		return
	}
	if method, ok := alias.Methods[name]; ok {
		expr.typing = method
		return
	}

	object, ok := alias.Ref.(Object)
	if !ok {
		p.report(
			fmt.Sprintf("Property '%v' doesn't exist on this type", name),
			expr.Property.Loc(),
		)
		expr.typing = Primitive{UNKNOWN}
		return
	}

	t, ok := object.Members[name]
	if !ok {
		p.report(
			fmt.Sprintf("Property '%v' doesn't exist on this type", name),
			expr.Property.Loc(),
		)
		expr.typing = Primitive{UNKNOWN}
		return
	}
	expr.typing = t
}

type TraitExpression struct {
	Receiver *ParenthesizedExpression // Receiver.Expr is an Identifier
	Def      *ParenthesizedExpression // contains *TupleExpression
}

func (t *TraitExpression) Loc() Loc {
	return Loc{t.Receiver.loc.Start, t.Def.loc.End}
}
func (t *TraitExpression) Type() ExpressionType {
	return Trait{
		Self:    Generic{Name: t.Receiver.Expr.(*Identifier).Text()},
		Members: t.Def.Type().(Type).Value.(Object).Members,
	}
}
func (t *TraitExpression) typeCheck(p *Parser) {
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()

	receiver := t.Receiver.Expr.(*Identifier)
	if receiver != nil {
		p.scope.Add(
			receiver.Text(),
			receiver.Loc(),
			Generic{Name: receiver.Text()},
		)
	}

	for _, element := range t.Def.Expr.(*TupleExpression).Elements {
		param, ok := element.(*Param)
		if !ok {
			continue
		}
		typing, ok := param.Complement.Type().(Type)
		if !ok || typing.Value == nil || typing.Value.Kind() != FUNCTION {
			p.report("Function type expected", param.Complement.Loc())
		}
	}
}

func parseTraitExpression(p *Parser, left Expression) Expression {
	paren := p.parseParenthesizedExpression()
	if paren != nil {
		paren.Expr = makeTuple(paren.Expr)
		validateFunctionParams(p, paren)
	}
	return &TraitExpression{
		Receiver: left.(*ParenthesizedExpression),
		Def:      paren,
	}
}
