package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TypedExpression struct {
	Expr     Node
	operator tokenizer.Token
	Typing   Node
}

func (t TypedExpression) Type() ExpressionType {
	// FIXME:
	return Primitive{UNKNOWN}
}
func (t TypedExpression) Check(c *Checker) {
	switch expr := t.Expr.(type) {
	case *TokenExpression:
		if expr.Token.Kind() != tokenizer.IDENTIFIER {
			c.report("Identifer expected", expr.Loc())
		}
	default:
		c.report("Identifer expected", expr.Loc())
	}

	t.Typing.Check(c)
	if t.Typing.Type().Kind() != TYPE {
		c.report("Type expected", t.Typing.Loc())
	}
}
func (t TypedExpression) Loc() tokenizer.Loc {
	loc := t.operator.Loc()
	if t.Expr != nil {
		loc.Start = t.Expr.Loc().Start
	}
	if t.Typing != nil {
		loc.End = t.Typing.Loc().End
	}
	return loc
}

func ParseTypedExpression(p *Parser) Expression {
	expr := ParseExpression(p)
	if p.tokenizer.Peek().Kind() != tokenizer.COLON {
		return expr
	}
	operator := p.tokenizer.Consume()
	typing := ParseExpression(p)
	return TypedExpression{expr, operator, typing}
}

func CheckTypedIdentifier(c *Checker, expr Expression) (string, bool) {
	typedExpression, ok := expr.(TypedExpression)
	if !ok {
		c.report("Typed identifier expected (name: type)", expr.Loc())
		return "", false
	}

	tokenExpression, ok := typedExpression.Expr.(*TokenExpression)
	if !ok {
		c.report("Identifier expected", typedExpression.Loc())
		return "", false
	}

	if tokenExpression.Token.Kind() != tokenizer.IDENTIFIER {
		c.report("Identifier expected", tokenExpression.Loc())
		return "", false
	}

	typedExpression.Typing.Check(c)

	return tokenExpression.Token.Text(), true
}
