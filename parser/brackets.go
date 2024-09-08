package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

// Expression between brackets, such as `[Type]`
type BracketedExpression struct {
	Expr Node
	loc  tokenizer.Loc
}

func (b BracketedExpression) Loc() tokenizer.Loc { return b.loc }
func (p *Parser) parseBracketedExpression() BracketedExpression {
	if p.tokenizer.Peek().Kind() != tokenizer.LBRACKET {
		panic("'[' expected!")
	}
	loc := p.tokenizer.Consume().Loc()

	outer := p.allowEmptyExpr
	p.allowEmptyExpr = true
	expr := ParseExpression(p)
	p.allowEmptyExpr = outer

	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.RBRACKET {
		p.report("']' expected", next.Loc())
		if expr != nil {
			loc.End = expr.Loc().End
		}
	} else {
		loc.End = p.tokenizer.Consume().Loc().End
	}

	return BracketedExpression{expr, loc}
}
