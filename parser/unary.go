package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type UnaryExpression struct {
	Token tokenizer.Token
	Expr  Node
}

func (u UnaryExpression) Loc() tokenizer.Loc {
	loc := u.Token.Loc()
	if u.Expr != nil {
		loc.End = u.Loc().End
	}
	return loc
}

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

func (p *Parser) parseUnaryExpression() Node {
	switch p.tokenizer.Peek().Kind() {
	case tokenizer.QUESTION_MARK:
		token := p.tokenizer.Consume()
		expr := parseInnerUnary(p)
		return UnaryExpression{token, expr}
	case tokenizer.LBRACKET:
		brackets := p.parseBracketedExpression()
		expr := parseInnerUnary(p)
		if function, ok := expr.(FunctionExpression); ok {
			function.TypeParams = &brackets
			return function
		}
		return ListTypeExpression{brackets, expr}
	default:
		return p.parseAccessExpression()
	}
}

func parseInnerUnary(p *Parser) Node {
	memBrace := p.allowBraceParsing
	memCall := p.allowCallExpr
	p.allowBraceParsing = false
	p.allowCallExpr = false
	expr := p.parseUnaryExpression()
	p.allowBraceParsing = memBrace
	p.allowCallExpr = memCall
	return expr
}
