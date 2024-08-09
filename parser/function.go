package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type FunctionExpression struct {
	Params   TupleExpression
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

func ParseFunctionExpression(p *Parser) Node {
	expr := parseTupleExpression(p)

	tuple, ok := expr.(TupleExpression)
	if !ok {
		return tuple
	}
	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.SLIM_ARR && next.Kind() != tokenizer.FAT_ARR {
		return tuple
	}
	operator := p.tokenizer.Consume()

	next = p.tokenizer.Peek()
	if next.Kind() == tokenizer.LBRACE {
		p.report("Expression expected", next.Loc())
	}

	res := FunctionExpression{tuple, operator, ParseExpression(p), nil}
	if operator.Kind() == tokenizer.FAT_ARR {
		res.Body = ParseBody(p)
	}
	return res
}
