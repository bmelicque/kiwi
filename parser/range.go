package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type RangeExpression struct {
	Left     Expression
	Right    Expression
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

func (r RangeExpression) Type() ExpressionType {
	var typing ExpressionType
	if r.Left != nil {
		typing = r.Left.Type()
	} else if r.Right != nil {
		typing = r.Right.Type()
	}
	return Range{typing}
}

func (r RangeExpression) Check(c *Checker) {
	var typing ExpressionType
	if r.Left != nil {
		r.Left.Check(c)
		typing = r.Left.Type()
	}
	if r.Right != nil {
		r.Right.Check(c)
		if typing == nil {
			typing = r.Right.Type()
		} else if !typing.Match(r.Right.Type()) {
			c.report("Types don't match", r.Loc())
		}
	}

	// FIXME:
	if typing != (Primitive{NUMBER}) {
		c.report("Number type expected", r.Loc())
	}

	if r.Operator.Kind() == tokenizer.RANGE_INCLUSIVE && r.Right == nil {
		c.report("Right operand expected", r.Operator.Loc())
	}
}

func ParseRange(p *Parser) Expression {
	token := p.tokenizer.Peek()

	var left Expression
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
