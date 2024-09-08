package parser

import "github.com/bmelicque/test-parser/tokenizer"

type ArrayType struct {
	Bracketed BracketedExpression
	Type      Node // Cannot be nil
}

func (a ArrayType) Loc() tokenizer.Loc {
	end := a.Bracketed.Loc().End
	if a.Type != nil {
		end = a.Type.Loc().End
	}
	return tokenizer.Loc{Start: a.Bracketed.loc.Start, End: end}
}

// Returns either a BracketedExpression or an ArrayType
func (p *Parser) parseArrayType() Node {
	brackets := p.parseBracketedExpression()

	var expr Node
	switch p.tokenizer.Peek().Kind() {
	case tokenizer.LBRACKET:
		expr = p.parseArrayType()
	case tokenizer.LPAREN:
		expr = p.parseParenthesizedExpression()
	default:
		old := p.allowEmptyExpr
		p.allowEmptyExpr = true
		p.parseTokenExpression()
		p.allowEmptyExpr = old
	}

	if expr == nil {
		return brackets
	}
	return ArrayType{brackets, expr}
}
