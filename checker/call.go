package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type CallExpression struct {
	Callee   Expression
	TypeArgs *TupleExpression
	Args     *TupleExpression
}

func (c CallExpression) Loc() tokenizer.Loc {
	loc := c.Args.loc
	if c.Callee != nil {
		loc.Start = c.Callee.Loc().Start
	}
	return loc
}

// FIXME:
func (c CallExpression) Type() ExpressionType {
	callee := c.Callee
	if callee == nil {
		return nil
	}

	if calleeType, ok := callee.Type().(Function); ok {
		return calleeType.Returned
	} else {
		return nil
	}
}

func (c *Checker) checkCallExpression(expr parser.CallExpression) Expression {
	callee := c.checkExpression(expr.Callee)
	if expr.Args == nil && expr.TypeArgs == nil {
		return callee
	}

	var typeArgs *TupleExpression
	if expr.TypeArgs != nil && expr.TypeArgs.Expr != nil {
		ex := c.checkExpression(expr.TypeArgs.Expr)
		if e, ok := ex.(TupleExpression); !ok {
			typeArgs = &e
		} else {
			typeArgs = &TupleExpression{[]Expression{ex}, ex.Loc()}
		}
	}

	var args *TupleExpression
	if expr.TypeArgs != nil && expr.Args.Expr != nil {
		ex := c.checkExpression(expr.Args.Expr)
		if e, ok := ex.(TupleExpression); !ok {
			args = &e
		} else {
			args = &TupleExpression{[]Expression{ex}, ex.Loc()}
		}
	}

	if callee.Type() != nil && callee.Type().Kind() == TYPE {
		// TODO: make sure callee is a generic type
		// TODO: check if number of args match
	} else {
		c.checkFunctionCallee(callee, typeArgs, args)
	}

	return CallExpression{callee, typeArgs, args}
}

func (c *Checker) checkFunctionCallee(callee Expression, typeArgs *TupleExpression, args *TupleExpression) {
	function, ok := callee.Type().(Function)
	if !ok {
		c.report("Function type expected", callee.Loc())
		return
	}

	c.pushScope(NewScope())
	defer c.dropScope()
	c.addTypeArgsToScope(typeArgs, function.TypeParams)

	// TODO: get type args if any

	// TODO: check if type args match number of type params
	// TODO: create good function type from type args
	// TODO: check if args match params form good function type
}

func (c *Checker) addTypeArgsToScope(args *TupleExpression, params []string) {
	var l int
	if args != nil {
		l = len(args.Elements)
	}

	if l > len(params) {
		loc := args.Elements[len(params)].Loc()
		loc.End = args.Elements[len(args.Elements)-1].Loc().End
		c.report("Too many type arguments", loc)
	}

	for i, param := range params {
		var arg Expression
		if i < l {
			arg = args.Elements[i]
		}
		if arg != nil {
			c.scope.Add(param, arg.Loc(), arg.Type())
		} else {
			c.scope.Add(param, tokenizer.Loc{}, Deferred{})
		}
	}
}
