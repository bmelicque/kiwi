package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func (c *Checker) addTypeArgsToScope(args *TupleExpression, params []Generic) {
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
		var loc tokenizer.Loc
		var t ExpressionType
		if i < l {
			arg := args.Elements[i]
			loc = arg.Loc()
			typing, ok := arg.Type().(Type)
			if ok {
				t = typing.Value
			} else {
				c.report("Typing expected", arg.Loc())
			}
		}
		if t != nil && param.Value != nil && !param.Value.Extends(t) {
			c.report("Type doesn't match", args.Elements[i].Loc())
		} else {
			params[i].Value = t
		}
		c.scope.Add(param.Name, loc, Type{Generic{Name: param.Name, Value: t}})
		v, _ := c.scope.Find(param.Name)
		v.reads = append(v.reads, loc)
	}
}

func addTypeParamsToScope(scope *Scope, params Params) {
	for _, param := range params.Params {
		if param.Typing == nil {
			name := param.Identifier.Text()
			t := Type{TypeAlias{Name: name, Ref: Generic{Name: name}}}
			scope.Add(name, param.Loc(), t)
		} else {
			// TODO: constrained generic
		}
	}
}

func checkBracketed(c *Checker, expr *parser.BracketedExpression) *TupleExpression {
	if expr == nil || expr.Expr == nil {
		return nil
	}
	ex := c.checkExpression(expr.Expr)
	if e, ok := ex.(TupleExpression); ok {
		return &e
	}
	return &TupleExpression{[]Expression{ex}, ex.Loc()}
}

func checkTypeIdentifier(c *Checker, node parser.Node) (Identifier, bool) {
	token, ok := node.(parser.TokenExpression)
	if !ok {
		return Identifier{}, false
	}

	identifier, ok := c.checkToken(token, false).(Identifier)
	if !ok {
		return Identifier{}, false
	}
	return identifier, identifier.isType
}
