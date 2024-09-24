package parser

import (
	"fmt"
	"slices"
	"unicode"

	"github.com/bmelicque/test-parser/tokenizer"
)

func ParseList(p *Parser, until tokenizer.TokenKind, callback func()) {
	multiline := false
	if p.tokenizer.Peek().Kind() == tokenizer.EOL {
		p.tokenizer.DiscardLineBreaks()
		multiline = true
	}

	next := p.tokenizer.Peek()
	for next.Kind() != until && next.Kind() != tokenizer.EOF {
		callback()

		exit := p.discardListBadTokens(until, multiline)
		if exit {
			return
		}

		if p.tokenizer.Peek().Kind() == until {
			return
		}

		next = p.tokenizer.Peek()
		if next.Kind() == tokenizer.COMMA {
			p.tokenizer.Consume()
		} else if !multiline && until != tokenizer.ILLEGAL {
			p.report("Expected ','", next.Loc())
		}

		next = p.tokenizer.Peek()
		if multiline && next.Kind() != tokenizer.EOL {
			p.report("Expected end of line", next.Loc())
		}
		if !multiline && next.Kind() == tokenizer.EOL {
			if until == tokenizer.ILLEGAL {
				return
			}
			p.report("Expected no end of line", next.Loc())
		}
		p.tokenizer.DiscardLineBreaks()

		next = p.tokenizer.Peek()
	}
}

// Return true if parseList should stop parsing
func (p *Parser) discardListBadTokens(until tokenizer.TokenKind, multiline bool) bool {
	next := p.tokenizer.Peek()
	var illegal tokenizer.Loc
	for next.Kind() != until && next.Kind() != tokenizer.COMMA && next.Kind() != tokenizer.EOL && next.Kind() != tokenizer.EOF {
		if until == tokenizer.ILLEGAL {
			return true
		}
		if illegal == (tokenizer.Loc{}) {
			illegal.Start = next.Loc().Start
		}
		illegal.End = next.Loc().End
		next = p.tokenizer.Peek()
	}
	if illegal != (tokenizer.Loc{}) {
		if multiline {
			p.report("Expected end of line", illegal)
		} else {
			p.report("Expected ','", illegal)
		}
	}
	return false
}

func recover(p *Parser, at tokenizer.TokenKind) bool {
	next := p.tokenizer.Peek()
	start := next.Loc().Start
	end := start
	recovery := []tokenizer.TokenKind{at, tokenizer.EOL, tokenizer.EOF}
	for ; slices.Contains(recovery, next.Kind()); next = p.tokenizer.Peek() {
		end = p.tokenizer.Consume().Loc().End
	}
	// FIXME: token text
	p.report(fmt.Sprintf("'%v' expected", at), tokenizer.Loc{Start: start, End: end})
	return next.Kind() == tokenizer.LBRACE
}

func IsTypeToken(expr Node) bool {
	token, ok := expr.(TokenExpression)
	if !ok {
		return false
	}

	if token.Token.Kind() != tokenizer.IDENTIFIER {
		return false
	}

	return unicode.IsUpper(rune(token.Token.Text()[0]))
}
