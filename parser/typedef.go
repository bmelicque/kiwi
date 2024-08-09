package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type ObjectDefinition struct {
	Members []Node
	loc     tokenizer.Loc
}

func (expr ObjectDefinition) Loc() tokenizer.Loc { return expr.loc }

func ParseObjectDefinition(p *Parser) Node {
	lbrace := p.tokenizer.Consume()
	loc := lbrace.Loc()

	members := []Node{}
	ParseList(p, tokenizer.RBRACE, func() {
		members = append(members, ParseTypedExpression(p))
	})

	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.RBRACE {
		p.report("'}' expected", next.Loc())
	}
	loc.End = p.tokenizer.Consume().Loc().End

	return ObjectDefinition{members, loc}
}
