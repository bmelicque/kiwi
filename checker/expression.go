package checker

import "github.com/bmelicque/test-parser/parser"

type ExpressionStatement struct {
	Expr Expression
}

func (e ExpressionStatement) Loc() parser.Loc { return e.Expr.Loc() }

func (c *Checker) checkExpressionStatement(expr parser.ExpressionStatement) ExpressionStatement {
	return ExpressionStatement{c.checkExpression(expr.Expr)}
}
