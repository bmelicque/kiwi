package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type CallExpression struct {
	Callee Expression
	Args   TupleExpression
	typing ExpressionType
}

func (c CallExpression) Loc() tokenizer.Loc {
	loc := c.Args.loc
	if c.Callee != nil {
		loc.Start = c.Callee.Loc().Start
	}
	return loc
}

func (c CallExpression) Type() ExpressionType { return c.typing }

func (c *Checker) checkCallExpression(expr parser.CallExpression) Expression {
	callee := c.checkExpression(expr.Callee)

	args := TupleExpression{loc: expr.Args.Loc()}
	if expr.Args.Expr != nil {
		ex := c.checkExpression(expr.Args.Expr)
		if e, ok := ex.(TupleExpression); ok {
			args = e
		} else {
			args = TupleExpression{[]Expression{ex}, ex.Loc()}
		}
	}

	returned := c.checkFunctionCallee(callee, &args)
	return CallExpression{callee, args, returned}
}

func (c *Checker) checkFunctionCallee(callee Expression, args *TupleExpression) ExpressionType {
	function, ok := callee.Type().(Function)
	if !ok {
		c.report("Function type expected", callee.Loc())
		return Primitive{UNKNOWN}
	}

	c.pushScope(NewScope())
	defer c.dropScope()
	for _, param := range function.TypeParams {
		// TODO: get declared location
		c.scope.Add(param.Name, tokenizer.Loc{}, Type{param})
	}

	params := function.Params.elements
	checkFunctionArgsNumber(c, args, params, callee.Loc())
	checkFunctionArgs(c, args, params)
	return function.Returned.build(c.scope, nil)
}

func checkFunctionArgsNumber(c *Checker, args *TupleExpression, params []ExpressionType, loc tokenizer.Loc) {
	if args == nil {
		c.report("Expected arguments", loc)
		return
	}

	if len(params) < len(args.Elements) {
		loc := args.Elements[len(params)].Loc()
		loc.End = args.Elements[len(args.Elements)-1].Loc().End
		c.report("Too many arguments", loc)
	}
	if len(params) > len(args.Elements) {
		c.report("Missing argument(s)", args.Loc())
	}
}

func checkFunctionArgs(c *Checker, args *TupleExpression, params []ExpressionType) {
	if args == nil {
		return
	}
	l := len(params)
	if len(args.Elements) < len(params) {
		l = len(args.Elements)
	}
	for i := 0; i < l; i++ {
		element := args.Elements[i]
		received := element.Type()
		params[i] = params[i].build(c.scope, received)
		if !params[i].Extends(received) {
			c.report("Types don't match", element.Loc())
		}
	}
}
