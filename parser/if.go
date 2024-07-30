package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type IfElse struct {
	keyword   tokenizer.Token
	Condition Expression
	Body      *Body
	// TODO: alternate
}

func (i IfElse) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: i.keyword.Loc().Start,
		End:   i.Body.Loc().End,
	}
	// TODO: handle alternate
}

func (i IfElse) Check(c *Checker) {
	if i.Condition != nil {
		i.Condition.Check(c)
		if i.Condition.Type(c.scope) != (Primitive{BOOLEAN}) {
			c.report("Boolean expected in condition", i.Condition.Loc())
		}
	}

	scope := NewScope()
	scope.returnType = c.scope.returnType
	// TODO: add variable to scope on some patterns:
	//		if Type {x} := y {}
	c.PushScope(scope)
	i.Body.Check(c)
	c.DropScope()
}

func ParseIfElse(p *Parser) Statement {
	keyword := p.tokenizer.Consume()
	// TODO: use ParseAssignment to check for pattern matching
	condition := ParseExpression(p)
	body := ParseBody(p)
	return IfElse{keyword, condition, body}
}
