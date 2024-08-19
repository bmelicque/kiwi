package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type RangeExpression struct {
	Left     Node
	Right    Node
	Operator tokenizer.Token
}

func (r RangeExpression) Loc() tokenizer.Loc {
	var loc tokenizer.Loc
	if r.Left != nil {
		loc.Start = r.Left.Loc().Start
	} else {
		loc.Start = r.Operator.Loc().Start
	}
	if r.Right != nil {
		loc.End = r.Right.Loc().End
	} else {
		loc.End = r.Operator.Loc().End
	}
	return loc
}

func ParseRange(p *Parser) Node {
	token := p.tokenizer.Peek()

	var left Node
	if token.Kind() != tokenizer.RANGE_INCLUSIVE && token.Kind() != tokenizer.RANGE_EXCLUSIVE {
		left = BinaryExpression{}.Parse(p)
	}

	token = p.tokenizer.Peek()
	if token.Kind() != tokenizer.RANGE_INCLUSIVE && token.Kind() != tokenizer.RANGE_EXCLUSIVE {
		return left
	}
	operator := p.tokenizer.Consume()

	right := BinaryExpression{}.Parse(p)

	return RangeExpression{left, right, operator}
}
