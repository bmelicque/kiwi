package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type Param struct {
	Identifier Identifier
	Typing     Expression // TODO: TypeExpression
}

func (p Param) Loc() tokenizer.Loc {
	var loc tokenizer.Loc
	if p.Identifier != (Identifier{}) {
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

func (p Param) Type() ExpressionType {
	if p.Typing == nil {
		return nil
	}
	typing, _ := p.Typing.Type().(Type)
	return typing.Value
}

func (c *Checker) checkParam(expr parser.TypedExpression) Param {
	var identifier Identifier
	if token, ok := expr.Expr.(parser.TokenExpression); ok {
		identifier, _ = c.checkToken(token, false).(Identifier)
	}
	if identifier == (Identifier{}) {
		c.report("Identifier expected", expr.Expr.Loc())
	}

	typing := c.checkExpression(expr.Typing)
	if _, ok := typing.Type().(Type); !ok {
		c.report("Typing expected", expr.Typing.Loc())
		typing = nil
	}

	return Param{identifier, typing}
}
