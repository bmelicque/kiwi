package checker

import (
	"github.com/bmelicque/test-parser/parser"
)

type CallExpression struct {
	Callee Expression
	Args   TupleExpression
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

func (c *Checker) checkCallExpression(expr parser.CallExpression) *CallExpression {
	callee := c.CheckExpression(expr.Callee)
	args, ok := c.CheckExpression(expr.Args).(TupleExpression)

	if !ok {
		c.report("Tuple expression expected", args.Loc())
		return nil
	}

	calleeType, ok := callee.Type().(Function)
	if !ok {
		c.report("Function type expected", callee.Loc())
		return nil
	}

	if !(args.Type().Extends(calleeType.params)) {
		c.report("Arguments types don't match expected parameters types", args.Loc())
	}

	return &CallExpression{callee, args}
}
