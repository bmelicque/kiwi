package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

func ParseList(p *Parser, until tokenizer.TokenKind, callback func()) {
	p.tokenizer.DiscardLineBreaks()
	next := p.tokenizer.Peek()
	for next.Kind() != until && next.Kind() != tokenizer.EOF {
		callback()

		// handle bad tokens before comma
		next = p.tokenizer.Peek()
		for next.Kind() != until && next.Kind() != tokenizer.COMMA && next.Kind() != tokenizer.EOF {
			p.report("',' expected", ParseExpression(p).Loc())
			next = p.tokenizer.Peek()
		}

		if next.Kind() == tokenizer.COMMA {
			p.tokenizer.Consume()
			p.tokenizer.DiscardLineBreaks()
		}
		next = p.tokenizer.Peek()
		if next.Kind() == until {
			break
		}
	}
}
