package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type TupleExpression struct {
	Elements []Expression
	loc      tokenizer.Loc
}

func (t TupleExpression) Type() ExpressionType {
	if len(t.Elements) == 0 {
		return Primitive{NIL}
	}
	if len(t.Elements) == 1 {
		return t.Elements[0].Type()
	}
	types := make([]ExpressionType, len(t.Elements))
	for i, element := range t.Elements {
		types[i] = element.Type()
	}
	return Tuple{types}
}

func (c *Checker) checkTuple(tuple parser.TupleExpression) TupleExpression {
	elements := make([]Expression, len(tuple.Elements))
	for i, element := range tuple.Elements {
		elements[i] = c.checkExpression(element)
	}
	return TupleExpression{elements, tuple.Loc()}
}

func (t TupleExpression) Loc() tokenizer.Loc {
	return t.loc
}
