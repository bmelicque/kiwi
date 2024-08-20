package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type FunctionExpression struct {
	Params   ParenthesizedExpression
	Operator tokenizer.Token // -> or =>
	Expr     Node            // return value for '->', return type for '=>'
	Body     *Body
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
	paren := p.parseParenthesizedExpression()
	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.SLIM_ARR && next.Kind() != tokenizer.FAT_ARR {
		return paren
	}
	operator := p.tokenizer.Consume()

	next = p.tokenizer.Peek()
	if next.Kind() == tokenizer.LBRACE {
		p.report("Expression expected", next.Loc())
	}

	res := FunctionExpression{paren, operator, ParseRange(p), nil}
	if operator.Kind() == tokenizer.FAT_ARR {
		res.Body = ParseBody(p)
	}
	return res
}
