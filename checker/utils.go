package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

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
			c.scope.Add(param, tokenizer.Loc{}, Type{Generic{Name: param}})
		}
	}
}

func checkTypeArgs(c *Checker, expr *parser.BracketedExpression) *TupleExpression {
	if expr == nil || expr.Expr == nil {
		return nil
	}
	ex := c.checkExpression(expr.Expr)
	if e, ok := ex.(TupleExpression); !ok {
		return &e
	}
	return &TupleExpression{[]Expression{ex}, ex.Loc()}
}
