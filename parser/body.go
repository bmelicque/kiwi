package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type Body struct {
	statements []Statement
	loc        tokenizer.Loc
}

func (b Body) Emit(e *Emitter) {
	e.Write("{")
	if len(b.statements) == 0 {
		e.Write("}")
		return
	}
	e.Write("\n")

	e.depth += 1
	for _, statement := range b.statements {
		e.Indent()
		statement.Emit(e)
	}
	e.depth -= 1

	e.Indent()
	e.Write("}\n")
}
func (b Body) Loc() tokenizer.Loc { return b.loc }
func (b Body) Check(c *Checker) {
	for _, statement := range b.statements {
		statement.Check(c)
	}
	// TODO: look for unreachable code
}

func ParseBody(p *Parser) *Body {
	body := Body{}

	token := p.tokenizer.Consume()
	body.loc.Start = token.Loc().Start
	if token.Kind() != tokenizer.LBRACE {
		p.report("'{' expected", token.Loc())
	}

	body.statements = []Statement{}
	for p.tokenizer.Peek().Kind() != tokenizer.RBRACE && p.tokenizer.Peek().Kind() != tokenizer.EOF {
		body.statements = append(body.statements, ParseStatement(p))
	}

	token = p.tokenizer.Consume()
	body.loc.End = token.Loc().End
	if token.Kind() != tokenizer.RBRACE {
		p.report("'}' expected", token.Loc())
	}

	return &body
}
