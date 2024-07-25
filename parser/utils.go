package parser

import (
	"unicode"

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

func IsType(expr Expression) bool {
	token, ok := expr.(TokenExpression)
	if !ok {
		return false
	}

	if token.Token.Kind() != tokenizer.IDENTIFIER {
		return false
	}

	return unicode.IsUpper(rune(token.Token.Text()[0]))
}