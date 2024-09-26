package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type UnaryExpression struct {
	Operator tokenizer.Token
	Operand  Expression
}

func (u UnaryExpression) Loc() tokenizer.Loc {
	loc := u.Operator.Loc()
	if u.Operand != nil {
		loc.End = u.Operand.Loc().End
	}
	return loc
}

func (u UnaryExpression) Type() ExpressionType {
	switch u.Operator.Kind() {
	case tokenizer.QUESTION_MARK:
		t := u.Operand.Type()
		if ty, ok := t.(Type); ok {
			t = ty.Value
		}
		return Type{makeOptionType(t)}
	default:
		return Primitive{UNKNOWN}
	}
}

func (c *Checker) checkUnaryExpression(node parser.UnaryExpression) UnaryExpression {
	switch node.Operator.Kind() {
	case tokenizer.QUESTION_MARK:
		operand := c.checkExpression(node.Operand)
		if operand.Type().Kind() != TYPE {
			c.report("Type expected", operand.Loc())
		}
		return UnaryExpression{node.Operator, operand}
	default:
		panic("This kind of unary expression is not implemented yet!")
	}
}
