package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type CallExpression struct {
	Callee Expression
	Args   TupleExpression
}

func (c CallExpression) Loc() tokenizer.Loc {
	loc := c.Args.loc
	if c.Callee != nil {
		loc.Start = c.Callee.Loc().Start
	}
	return loc
}

func (c CallExpression) Type() ExpressionType {
	callee := c.Callee
	if callee == nil {
		return nil
	}

	if calleeType, ok := callee.Type().(Function); ok {
		return calleeType.returned
	} else {
		return nil
	}
}

func (c *Checker) checkCallExpression(expr parser.CallExpression) CallExpression {
	callee := c.checkExpression(expr.Callee)
	paren, ok := c.checkExpression(expr.Args).(ParenthesizedExpression)
	if !ok {
		c.report("Expected arguments", expr.Args.Loc())
	}
	args, ok := paren.Expr.(TupleExpression)
	if !ok {
		args = TupleExpression{
			Elements: []Expression{paren.Expr},
			loc:      paren.Expr.Loc(),
		}
	}
	calleeType, ok := callee.Type().(Function)
	if !ok {
		c.report("Function type expected", callee.Loc())
	} else if !(calleeType.params.Extends(args.Type())) {
		c.report("Arguments types don't match expected parameters types", expr.Args.Loc())
	}

	return CallExpression{callee, args}
}
