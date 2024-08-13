package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type ExpressionStatement struct {
	Expr Expression
}

func (e ExpressionStatement) Loc() tokenizer.Loc { return e.Expr.Loc() }

func (c *Checker) checkExpressionStatement(expr parser.ExpressionStatement) ExpressionStatement {
	return ExpressionStatement{c.CheckExpression(expr.Expr)}
}
