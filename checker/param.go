package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type Param struct {
	Identifier *Identifier
	Typing     Expression // TODO: TypeExpression
}

func (p Param) Loc() tokenizer.Loc {
	var loc tokenizer.Loc
	if p.Identifier != nil {
		loc.Start = p.Identifier.Loc().Start
	} else {
		loc.Start = p.Typing.Loc().Start
	}

	if p.Typing != nil {
		loc.End = p.Typing.Loc().End
	} else {
		loc.End = p.Identifier.Loc().End
	}

	return loc
}

func (p Param) Type() ExpressionType { return nil }

func (c *Checker) checkParam(expr parser.TypedExpression) Param {
	var identifier *Identifier
	if token, ok := expr.Expr.(*parser.TokenExpression); ok {
		identifier, _ = c.checkToken(token).(*Identifier)
	}
	if identifier == nil {
		c.report("Identifier expected", expr.Expr.Loc())
	}

	typing := c.CheckExpression(expr.Typing)
	if _, ok := typing.Type().(Type); !ok {
		c.report("Typing expected", expr.Typing.Loc())
		typing = nil
	}

	return Param{identifier, typing}
}
