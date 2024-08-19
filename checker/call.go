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
	args, ok := c.checkExpression(expr.Args).(TupleExpression)
	if !ok {
		c.report("Tuple expression expected", args.Loc())
	}
	calleeType, ok := callee.Type().(Function)
	if !ok {
		c.report("Function type expected", callee.Loc())
	} else if !(calleeType.params.Extends(args.Type())) {
		c.report("Arguments types don't match expected parameters types", args.Loc())
	}

	return CallExpression{callee, args}
}
