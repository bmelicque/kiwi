package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type ParenthesizedExpression struct {
	Expr Node
	loc  tokenizer.Loc
}

func (p ParenthesizedExpression) Loc() tokenizer.Loc {
	return p.loc
}
func (p ParenthesizedExpression) Unwrap() Node {
	if expr, ok := p.Expr.(ParenthesizedExpression); ok {
		return expr.Unwrap()
	}
	return p.Expr
}

func (p *Parser) parseParenthesizedExpression() ParenthesizedExpression {
	loc := p.tokenizer.Consume().Loc() // LPAREN
	p.tokenizer.DiscardLineBreaks()
	next := p.tokenizer.Peek()
	if next.Kind() == tokenizer.RPAREN {
		loc.End = p.tokenizer.Consume().Loc().End
		return ParenthesizedExpression{nil, loc}
	}

	outerBrace := p.allowBraceParsing
	outerMultiline := p.multiline
	p.allowBraceParsing = true
	p.multiline = true
	expr := p.parseTupleExpression()
	p.allowBraceParsing = outerBrace
	p.multiline = outerMultiline

	p.tokenizer.DiscardLineBreaks()
	next = p.tokenizer.Peek()
	if next.Kind() != tokenizer.RPAREN {
		p.report("')' expected", next.Loc())
	}
	loc.End = p.tokenizer.Consume().Loc().End
	return ParenthesizedExpression{expr, loc}
}

// unwrap parenthesized expressions
func Unwrap(node Node) Node {
	if paren, ok := node.(ParenthesizedExpression); ok {
		return paren.Unwrap()
	}
	return node
}
