package checker

import (
	"fmt"
	"strconv"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type PropertyAccessExpression struct {
	Expr     Expression
	Property Identifier
	typing   ExpressionType
}

func (p PropertyAccessExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: p.Expr.Loc().Start,
		End:   p.Property.Loc().End,
	}
}
func (p PropertyAccessExpression) Type() ExpressionType { return p.typing }

type TraitExpression struct {
	Receiver ParenthesizedExpression // Receiver.Expr is an Identifier
	Def      ObjectDefinition
}

func (t TraitExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: t.Receiver.loc.Start,
		End:   t.Def.loc.End,
	}
}
func (t TraitExpression) Type() ExpressionType {
	value := Trait{
		Self:    Generic{Name: t.Receiver.Expr.(Identifier).Text()},
		Members: map[string]ExpressionType{},
	}
	for _, member := range t.Def.Members {
		value.Members[member.Name.Token.Text()] = member.Typing.Type()
	}
	return Type{value}
}

// Returns the type of the given object as a `TypeRef{Object{}}`
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
func (c *Checker) checkPropertyAccess(expr parser.PropertyAccessExpression) Expression {
	left := c.checkExpression(expr.Expr)
	if _, ok := expr.Property.(parser.ObjectDefinition); ok {
		trait := checkTraitExpression(c, left, expr.Property)
		if trait != nil {
			return trait
		}
	}
	switch left.Type().(type) {
	case Tuple:
		return checkTupleIndexAccess(c, left, expr.Property)
	case Type:
		return checkSumTypeConstructorAccess(c, left, expr.Property)
	default:
		return checkObjectPropertyAccess(c, left, expr.Property)
	}
}

// check accessing an object's property or method: object.property
func checkObjectPropertyAccess(c *Checker, left Expression, right parser.Node) PropertyAccessExpression {
	property, ok := c.checkExpression(right).(Identifier)
	if !ok {
		c.report("Identifier expected", right.Loc())
	}
	name := property.Token.Text()

	object, ok := getObjectType(left)
	if !ok {
		c.report("Object type expected", left.Loc())
	}

	var typing ExpressionType
	if method, ok := c.scope.FindMethod(name, object); ok {
		typing = method.signature
	} else {
		typing = object.Ref.(Object).Members[name].(Type).Value
	}

	return PropertyAccessExpression{
		Expr:     left,
		Property: property,
		typing:   typing,
	}
}

// check accessing a tuple's index: tuple.0
func checkTupleIndexAccess(c *Checker, left Expression, right parser.Node) PropertyAccessExpression {
	property, ok := c.checkExpression(right).(Literal)
	if !ok || property.Type().Kind() != NUMBER {
		c.report("Number expected", right.Loc())
	}
	typing := getTupleAccessType(c, property, left.Type())

	return PropertyAccessExpression{
		Expr:     left,
		Property: Identifier{right.(parser.TokenExpression), property.Type(), false}, // FIXME:
		typing:   typing,
	}
}
func getTupleAccessType(c *Checker, property Literal, typing ExpressionType) ExpressionType {
	if property.Type().Kind() != NUMBER {
		return Primitive{UNKNOWN}
	}
	number, err := strconv.Atoi(property.Text())
	if err != nil {
		c.report("Integer expected", property.Loc())
		return Primitive{UNKNOWN}
	}
	if number > len(typing.(Tuple).elements)-1 || number < 0 {
		c.report("Index out of range", property.Loc())
		return Primitive{UNKNOWN}
	}
	return typing.(Tuple).elements[number]
}

// check accessing a sum type's subconstructor: SumType.Constructor
func checkSumTypeConstructorAccess(c *Checker, left Expression, right parser.Node) PropertyAccessExpression {
	property, ok := c.checkExpression(right).(Identifier)
	if !ok {
		c.report("Identifier expected", right.Loc())
	}
	name := property.Token.Text()

	typing := getSumTypeConstructor(left.Type().(Type), name)
	if typing == (Primitive{UNKNOWN}) {
		c.report(fmt.Sprintf("Property '%v' doesn't exist on this type", name), right.Loc())
	}
	return PropertyAccessExpression{
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

// check a trait expression: (ReceiverType).{ ..methods }
func checkTraitExpression(c *Checker, left Expression, right parser.Node) Expression {
	receiver, ok := left.(ParenthesizedExpression)
	if !ok {
		return nil
	}
	identifier, ok := receiver.Expr.(Identifier)
	if !ok {
		return nil
	}
	if !identifier.isType {
		c.report("Type expected", identifier.Loc())
	}

	c.pushScope(NewScope())
	defer c.dropScope()
	name := identifier.Text()
	c.scope.Add(name, identifier.Loc(), Type{TypeAlias{Name: name, Ref: Generic{Name: identifier.Text()}}})

	def := c.checkObjectDefinition(right.(parser.ObjectDefinition)) // ensured by checkPropertyAccess
	for _, member := range def.Members {
		typing, _ := member.Typing.Type().(Type)
		if typing.Value == nil || typing.Value.Kind() != FUNCTION {
			c.report("Function type expected", member.Typing.Loc())
		}
	}

	return TraitExpression{
		Receiver: receiver,
		Def:      def,
	}
}
