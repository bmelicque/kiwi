package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type Params struct {
	Params []Param
	loc    tokenizer.Loc
}

func (p Params) Loc() tokenizer.Loc { return p.loc }
func (p Params) Type() ExpressionType {
	if len(p.Elements) == 0 {
		return Primitive{NIL}
	}
	if len(p.Elements) == 1 {
		return p.Elements[0].Type()
	}
	types := make([]ExpressionType, len(p.Elements))
	for i, element := range p.Elements {
		types[i] = element.Type()
	}
	return Tuple{types}
}

func (c *Checker) checkParams(params parser.TupleExpression) Params {
	elements := make([]Param, len(params.Elements))
	for i, element := range params.Elements {
		if param, ok := element.(parser.TypedExpression); ok {
			elements[i] = c.checkParam(param)
		}
	}
	return Params{elements, params.Loc()}
}
