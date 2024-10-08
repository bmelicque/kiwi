package checker

import (
	"fmt"
	"unicode"

	"github.com/bmelicque/test-parser/parser"
)

type Literal struct {
	parser.TokenExpression
}

func (l Literal) Type() ExpressionType {
	switch l.Token.Kind() {
	case parser.NumberLiteral:
		return Primitive{NUMBER}
	case parser.BooleanLiteral:
		return Primitive{BOOLEAN}
	case parser.StringLiteral:
		return Primitive{STRING}
	case parser.StringKeyword:
		return Type{Primitive{STRING}}
	case parser.NumberKeyword:
		return Type{Primitive{NUMBER}}
	case parser.BooleanKeyword:
		return Type{Primitive{BOOLEAN}}
	default:
		panic(fmt.Sprintf("Unknown typing kind: %v (not implemented yet)", l.Token.Kind()))
	}
}

type Identifier struct {
	parser.TokenExpression
	typing ExpressionType
	isType bool
}

func (i Identifier) Loc() parser.Loc      { return i.TokenExpression.Loc() }
func (i Identifier) Type() ExpressionType { return i.typing }

func (c *Checker) checkToken(t parser.TokenExpression, report bool) Expression {
	if t.Token.Kind() != parser.Name {
		return Literal{t}
	}

	name := t.Token.Text()
	isType := unicode.IsUpper(rune(name[0]))
	if !report {
		return Identifier{t, nil, isType}
	}

	var typing ExpressionType
	if variable, ok := c.scope.Find(name); ok {
		c.scope.ReadAt(name, t.Loc())
		typing = variable.typing
	}
	return Identifier{t, typing, isType}
}
