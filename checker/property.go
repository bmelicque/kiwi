package checker

import (
	"fmt"

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
func (c *Checker) checkPropertyAccess(expr parser.PropertyAccessExpression) PropertyAccessExpression {
	left := c.checkExpression(expr.Expr)
	property, ok := c.checkExpression(expr.Property).(Identifier)
	if !ok {
		c.report("Expected identifier", expr.Property.Loc())
	}
	name := property.Token.Text()

	if t, ok := left.Type().(Type); ok {
		typing := checkSumTypeConstructor(t, name)
		if typing == (Primitive{UNKNOWN}) {
			c.report(fmt.Sprintf("Property '%v' doesn't exist on this type", name), expr.Loc())
		}
		return PropertyAccessExpression{
			Expr:     left,
			Property: property,
			typing:   typing,
		}
	}

	object, ok := getObjectType(left)
	if !ok {
		c.report("Expected object type", expr.Expr.Loc())
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

func checkSumTypeConstructor(t Type, name string) ExpressionType {
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
	return constructor
}
