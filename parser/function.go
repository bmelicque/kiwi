package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type FunctionExpression struct {
	TypeParams *BracketedExpression
	Params     *ParenthesizedExpression
	Operator   tokenizer.Token // -> or =>
	Expr       Node            // return value for '->', return type for '=>'
	Body       *Body
}

func (f FunctionExpression) Loc() tokenizer.Loc {
	loc := tokenizer.Loc{Start: f.Params.Loc().Start, End: tokenizer.Position{}}
	if f.Body == nil {
		loc.End = f.Expr.Loc().End
	} else {
		loc.End = f.Body.Loc().End
	}
	return loc
}

func (p *Parser) parseFunctionExpression() Node {
	var brackets *BracketedExpression
	if p.tokenizer.Peek().Kind() == tokenizer.LBRACKET {
		b := p.parseBracketedExpression()
		brackets = &b
	}
	var paren *ParenthesizedExpression
	if p.tokenizer.Peek().Kind() == tokenizer.LPAREN {
		pa := p.parseParenthesizedExpression()
		paren = &pa
	}

	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.SLIM_ARR && next.Kind() != tokenizer.FAT_ARR {
		// FIXME: return either angle, paren or typed(angle, paren)
		return *paren
	}
	operator := p.tokenizer.Consume()

	next = p.tokenizer.Peek()
	if next.Kind() == tokenizer.LBRACE {
		p.report("Expression expected", next.Loc())
	}

	res := FunctionExpression{brackets, paren, operator, ParseRange(p), nil}
	if operator.Kind() == tokenizer.FAT_ARR {
		res.Body = ParseBody(p)
	}
	return res
}
