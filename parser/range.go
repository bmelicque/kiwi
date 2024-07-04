package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type RangeExpression struct {
	left     Expression
	right    Expression
	operator tokenizer.Token
}

func (r RangeExpression) Loc() tokenizer.Loc {
	var loc tokenizer.Loc
	if r.left != nil {
		loc.Start = r.left.Loc().Start
	} else {
		loc.Start = r.operator.Loc().Start
	}
	if r.right != nil {
		loc.End = r.right.Loc().End
	} else {
		loc.End = r.operator.Loc().End
	}
	return loc
}

// TODO: Emit
func (r RangeExpression) Emit(e *Emitter) {
	e.AddFlag(RangeFlag)

	e.Write("range(")

	if r.left != nil {
		r.left.Emit(e)
	} else {
		e.Write("0")
	}

	e.Write(", ")

	if r.right != nil {
		r.right.Emit(e)
		if r.operator.Kind() == tokenizer.RANGE_INCLUSIVE {
			e.Write(" + 1")
		}
	} else {
		e.Write("1")
	}

	e.Write(")")
}

func (r RangeExpression) Type(ctx *Scope) ExpressionType {
	var typing ExpressionType
	if r.left != nil {
		typing = r.left.Type(ctx)
	} else if r.right != nil {
		typing = r.right.Type(ctx)
	}
	return Range{typing}
}

func (r RangeExpression) Check(c *Checker) {
	var typing ExpressionType
	if r.left != nil {
		r.left.Check(c)
		typing = r.left.Type(c.scope)
	}
	if r.right != nil {
		r.right.Check(c)
		if typing == nil {
			typing = r.right.Type(c.scope)
		} else if !typing.Match(r.right.Type(c.scope)) {
			c.report("Types don't match", r.Loc())
		}
	}

	// FIXME:
	if typing != (Primitive{NUMBER}) {
		c.report("Number type expected", r.Loc())
	}

	if r.operator.Kind() == tokenizer.RANGE_INCLUSIVE && r.right == nil {
		c.report("Right operand expected", r.operator.Loc())
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
