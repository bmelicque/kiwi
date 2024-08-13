package checker

import (
	"fmt"
	"unicode"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type Literal struct {
	parser.TokenExpression
}

func (l Literal) Type() ExpressionType {
	switch l.Token.Kind() {
	case tokenizer.NUMBER:
		return Primitive{NUMBER}
	case tokenizer.BOOLEAN:
		return Primitive{BOOLEAN}
	case tokenizer.STRING:
		return Primitive{STRING}
	case tokenizer.STR_KW:
		return Type{Primitive{STRING}}
	case tokenizer.NUM_KW:
		return Type{Primitive{NUMBER}}
	case tokenizer.BOOL_KW:
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

func (i Identifier) Loc() tokenizer.Loc   { return i.TokenExpression.Loc() }
func (l Identifier) Type() ExpressionType { return l.typing }

func (c *Checker) checkToken(t *parser.TokenExpression, report bool) Expression {
	if t.Token.Kind() != tokenizer.IDENTIFIER {
		return Literal{*t}
	}

	name := t.Token.Text()
	isType := unicode.IsUpper(rune(name[0]))
	if !report {
		return Identifier{*t, nil, isType}
	}

	var typing ExpressionType
	if variable, ok := c.scope.Find(name); ok {
		c.scope.ReadAt(name, t.Loc())
		typing = variable.typing
	}
	if isType {
		typing = TypeRef{name, typing.(Type).Value}
		isType = true
	}
	return Identifier{*t, typing, isType}
}
