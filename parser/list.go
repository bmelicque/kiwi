package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type ListTypeExpression struct {
	Bracketed BracketedExpression
	Type      Node // Cannot be nil
}

func (l ListTypeExpression) Loc() tokenizer.Loc {
	end := l.Bracketed.Loc().End
	if l.Type != nil {
		end = l.Type.Loc().End
	}
	return tokenizer.Loc{Start: l.Bracketed.loc.Start, End: end}
}

// Returns either a BracketedExpression or an ArrayType
func (p *Parser) parseListTypeExpression() Node {
	brackets := p.parseBracketedExpression()

	var expr Node
	switch p.tokenizer.Peek().Kind() {
	case tokenizer.LBRACKET:
		expr = p.parseListTypeExpression()
	case tokenizer.LPAREN:
		expr = p.parseParenthesizedExpression()
	default:
		old := p.allowEmptyExpr
		p.allowEmptyExpr = true
		expr = p.parseTokenExpression()
		p.allowEmptyExpr = old
	}

	if expr == nil {
		return brackets
	}
	return ListTypeExpression{brackets, expr}
}
