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

func (p PropertyAccessExpression) Loc() Loc {
	return Loc{
		Start: p.Expr.Loc().Start,
		End:   p.Property.Loc().End,
	}
}
func (p PropertyAccessExpression) Type() ExpressionType { return p.typing }

func parsePropertyAccess(p *Parser, left Expression) Expression {
	prop := fallback(p)
	if _, ok := prop.(*ParenthesizedExpression); ok {
		trait := getValidatedTraitExpression(p, left, prop)
		if trait != nil {
			return trait
		}
	}
	switch left.Type().(type) {
	case Tuple:
		return getValidatedTupleIndexAccess(p, left, prop)
	case Type:
		return getValidatedSumConstructorAccess(p, left, prop)
	default:
		return getValidatedObjectPropertyAccess(p, left, prop)
	}
}

// check accessing a tuple's index: tuple.0
func getValidatedTupleIndexAccess(p *Parser, left Expression, right Expression) *PropertyAccessExpression {
	property, ok := right.(*Literal)
	if !ok || property.Type().Kind() != NUMBER {
		p.report("Number expected", right.Loc())
	}
	typing := getTupleAccessType(p, *property, left.Type())

	return &PropertyAccessExpression{
		Expr:     left,
		Property: property,
		typing:   typing,
	}
}
func getTupleAccessType(p *Parser, property Literal, typing ExpressionType) ExpressionType {
	if property.Type().Kind() != NUMBER {
		return Primitive{UNKNOWN}
	}
	number, err := strconv.Atoi(property.Text())
	if err != nil {
		p.report("Integer expected", property.Loc())
		return Primitive{UNKNOWN}
	}
	if number > len(typing.(Tuple).elements)-1 || number < 0 {
		p.report("Index out of range", property.Loc())
		return Primitive{UNKNOWN}
	}
	return typing.(Tuple).elements[number]
}

// check accessing a sum type's subconstructor: SumType.Constructor
func getValidatedSumConstructorAccess(p *Parser, left Expression, right Expression) *PropertyAccessExpression {
	property, ok := right.(*Identifier)
	if !ok {
		p.report("Identifier expected", right.Loc())
	}
	name := property.Token.Text()

	typing := getSumTypeConstructor(left.Type().(Type), name)
	if typing == (Primitive{UNKNOWN}) {
		p.report(fmt.Sprintf("Property '%v' doesn't exist on this type", name), right.Loc())
	}
	return &PropertyAccessExpression{
		Expr:     left,
		Property: property,
		typing:   typing,
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
	return Type{Constructor{
		From: constructor,
		To:   t,
	}}
}

// check accessing an object's property or method: object.property
func getValidatedObjectPropertyAccess(p *Parser, left Expression, right Expression) *PropertyAccessExpression {
	property, ok := right.(*Identifier)
	if right != nil && !ok {
		p.report("Identifier expected", right.Loc())
	}
	var name string
	if property != nil {
		name = property.Token.Text()
	}

	object, ok := getObjectType(left)
	if !ok {
		p.report("Object type expected", left.Loc())
		return &PropertyAccessExpression{
			Expr:     left,
			Property: property,
		}
	}

	// FIXME: methods should be on the TypeAlias
	var typing ExpressionType
	if method, ok := p.scope.FindMethod(name, object); ok {
		typing = method.signature
	} else {
		typing = object.Ref.(Object).Members[name].(Type).Value
	}

	return &PropertyAccessExpression{
		Expr:     left,
		Property: property,
		typing:   typing,
	}
}

// Return the type of the given object as a `TypeAlias{Object{}}`
func getObjectType(expr Expression) (TypeAlias, bool) {
	ref, ok := expr.Type().(TypeAlias)
	if !ok {
		return TypeAlias{}, false
	}
	if _, ok := ref.Ref.(Object); !ok {
		return TypeAlias{}, false
	}
	return ref, true
}

type TraitExpression struct {
	Receiver *ParenthesizedExpression // Receiver.Expr is an Identifier
	Def      *ParenthesizedExpression
}

func (t TraitExpression) Loc() Loc {
	return Loc{t.Receiver.loc.Start, t.Def.loc.End}
}
func (t TraitExpression) Type() ExpressionType {
	return Trait{
		Self:    Generic{Name: t.Receiver.Expr.(Identifier).Text()},
		Members: t.Def.Type().(Type).Value.(Object).Members,
	}
}

// check a trait expression: (ReceiverType).{ ..methods }
func getValidatedTraitExpression(p *Parser, left Expression, right Expression) Expression {
	receiver, ok := left.(*ParenthesizedExpression)
	if !ok {
		return nil
	}
	identifier, ok := receiver.Expr.(*Identifier)
	if !ok {
		return nil
	}
	if !identifier.isType {
		p.report("Type expected", identifier.Loc())
	}

	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	name := identifier.Text()
	p.scope.Add(name, identifier.Loc(), Type{TypeAlias{Name: name, Ref: Generic{Name: identifier.Text()}}})

	paren := right.(*ParenthesizedExpression)
	validateTraitType(p, *paren)

	return TraitExpression{
		Receiver: receiver,
		Def:      paren,
	}
}

func validateTraitType(p *Parser, expr ParenthesizedExpression) {
	ty, ok := expr.Type().(Type)
	if !ok {
		p.report("Object type expected", expr.Loc())
		return
	}
	if _, ok := ty.Value.(Object); !ok {
		p.report("Object type expected", expr.Loc())
		return
	}

	tuple, ok := expr.Expr.(*TupleExpression)
	if !ok {
		validateTraitMethod(p, expr.Expr)
		return
	}
	for _, element := range tuple.Elements {
		validateTraitMethod(p, element)
	}
}

func validateTraitMethod(p *Parser, expr Expression) {
	param, ok := expr.(Param)
	if !ok {
		p.report("Method declaration expected", expr.Loc())
		return
	}
	typing, ok := param.Complement.Type().(Type)
	if !ok || typing.Value == nil || typing.Value.Kind() != FUNCTION {
		p.report("Function type expected", param.Complement.Loc())
	}
}
