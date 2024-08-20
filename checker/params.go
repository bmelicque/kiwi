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
	types := make([]ExpressionType, len(p.Params))
	for i, element := range p.Params {
		types[i] = element.Type()
	}
	return Tuple{types}
}

func (c *Checker) checkParams(params parser.ParenthesizedExpression) Params {
	tuple, ok := params.Expr.(parser.TupleExpression)
	if !ok {
		param, ok := params.Expr.(parser.TypedExpression)
		if !ok {
			c.report("Expected typed identifier", param.Expr.Loc())
			return Params{}
		}
		return Params{[]Param{c.checkParam(param)}, params.Loc()}
	}
	elements := make([]Param, len(tuple.Elements))
	for i, element := range tuple.Elements {
		if param, ok := element.(parser.TypedExpression); ok {
			elements[i] = c.checkParam(param)
		}
	}
	return Params{elements, params.Loc()}
}
