package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type Node interface {
	Loc() tokenizer.Loc
}

func fallback(p *Parser) Node {
	switch p.tokenizer.Peek().Kind() {
	case tokenizer.LBRACKET:
		return p.parseUnaryExpression()
	case tokenizer.LPAREN:
		return p.parseFunctionExpression()
	case tokenizer.LBRACE:
		if p.allowBraceParsing {
			return p.parseObjectDefinition()
		}
	}
	return p.parseTokenExpression()
}
