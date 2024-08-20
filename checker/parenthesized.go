package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type ParenthesizedExpression struct {
	Expr Expression
	loc  tokenizer.Loc
}

func (p ParenthesizedExpression) Loc() tokenizer.Loc { return p.loc }
func (p ParenthesizedExpression) Type() ExpressionType {
	if p.Expr == nil {
		return Primitive{NIL}
	}
	return p.Expr.Type()
}

func (c *Checker) checkParenthesizedExpression(expr parser.ParenthesizedExpression) ParenthesizedExpression {
	return ParenthesizedExpression{
		Expr: c.checkExpression(expr.Expr),
		loc:  expr.Loc(),
	}
}

// TODO: Check
