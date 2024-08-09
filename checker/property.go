package checker

import (
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
func getObjectType(expr Expression) (TypeRef, bool) {
	ref, ok := expr.Type().(TypeRef)
	if !ok {
		return TypeRef{}, false
	}
	if _, ok := ref.Ref.(Object); !ok {
		return TypeRef{}, false
	}
	return ref, true
}
func (c *Checker) checkPropertyAccess(expr parser.PropertyAccessExpression) PropertyAccessExpression {
	left := c.CheckExpression(expr.Expr)
	object, ok := getObjectType(left)
	if !ok {
		c.report("Expected object type", expr.Expr.Loc())
	}

	property, ok := c.CheckExpression(expr.Property).(Identifier)
	if !ok {
		c.report("Expected identifier", expr.Property.Loc())
	}
	name := property.Token.Text()

	var typing ExpressionType
	if method, ok := c.scope.FindMethod(name, object); ok {
		typing = method.signature
	} else {
		typing = object.Ref.(Object).Members[name]
	}

	return PropertyAccessExpression{
		Expr:     left,
		Property: property,
		typing:   typing,
	}
}
