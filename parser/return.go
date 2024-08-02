package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type Return struct {
	operator tokenizer.Token
	Value    Expression
}

func (r Return) Loc() tokenizer.Loc {
	loc := r.operator.Loc()
	if r.Value != nil {
		loc.End = r.Value.Loc().End
	}
	return loc
}

func (r Return) Check(c *Checker) {
	expected := c.scope.returnType

	if r.Value != nil {
		r.Value.Check(c)

		if expected == nil {
			c.report("Expected no return value", r.Loc())
		} else if !expected.Match(r.Value.Type()) {
			c.report("Type does not match expected return type", r.Value.Loc())
		}
	} else if expected != nil {
		c.report("Expected return value", r.Loc())

	}
}

func ParseReturn(p *Parser) Statement {
	keyword := p.tokenizer.Consume()

	if p.tokenizer.Peek().Kind() == tokenizer.EOL {
		return Return{keyword, nil}
	}

	value := ParseExpression(p)
	return Return{keyword, value}
}
