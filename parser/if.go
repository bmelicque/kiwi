package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type IfElse struct {
	keyword   tokenizer.Token
	condition Expression
	body      *Body
	// TODO: alternate
}

// TODO: handle alternate
func (i IfElse) Emit(e *Emitter) {
	e.Write("if (")
	i.condition.Emit(e)
	e.Write(")")

	e.Write(" ")
	i.body.Emit(e)
}

func (i IfElse) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: i.keyword.Loc().Start,
		End:   i.body.Loc().End,
	}
	// TODO: handle alternate
}

func (i IfElse) Check(c *Checker) {
	if i.condition != nil {
		i.condition.Check(c)
		if i.condition.Type(c.scope) != (Primitive{BOOLEAN}) {
			c.report("Boolean expected in condition", i.condition.Loc())
		}
	}

	scope := Scope{map[string]*Variable{}, nil, nil}
	scope.returnType = c.scope.returnType
	// TODO: add variable to scope on some patterns:
	//		if Type {x} := y {}
	c.PushScope(&scope)
	i.body.Check(c)
	c.DropScope()
}

func ParseIfElse(p *Parser) Statement {
	keyword := p.tokenizer.Consume()
	// TODO: use ParseAssignment to check for pattern matching
	condition := ParseExpression(p)
	body := ParseBody(p)
	return IfElse{keyword, condition, body}
}
