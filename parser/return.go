package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type Return struct {
	operator tokenizer.Token
	value    Expression
}

func (r Return) Emit(e *Emitter) {
	e.Write("return")
	if r.value != nil {
		e.Write(" ")
		r.value.Emit(e)
	}
	e.Write(";\n")
}

func (r Return) Loc() tokenizer.Loc {
	loc := r.operator.Loc()
	if r.value != nil {
		loc.End = r.value.Loc().End
	}
	return loc
}

func (r Return) Check(c *Checker) {
	expected := c.scope.returnType

	if r.value != nil {
		r.value.Check(c)

		if expected == nil {
			c.report("Expected no return value", r.Loc())
		} else if !expected.Match(r.value.Type(c.scope)) {
			c.report("Type does not match expected return type", r.value.Loc())
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
