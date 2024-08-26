package checker

import (
	"unicode"

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

func (c *Checker) checkParams(params parser.Node) Params {
	tuple, ok := params.(parser.TupleExpression)
	if !ok {
		return c.checkSingleParams(params)
	}
	elements := make([]Param, len(tuple.Elements))
	for i, element := range tuple.Elements {
		if param, ok := element.(parser.TypedExpression); ok {
			elements[i] = c.checkParam(param)
		}
	}
	return Params{elements, params.Loc()}
}

func (c *Checker) checkSingleParams(params parser.Node) Params {
	switch param := params.(type) {
	case parser.TypedExpression:
		return Params{[]Param{c.checkParam(param)}, params.Loc()}
	case parser.TokenExpression:
		identifier, ok := c.checkToken(param, false).(Identifier)
		if !ok {
			break
		}
		if !unicode.IsUpper(rune(identifier.Token.Text()[0])) {
			c.report("Expected a type identifier", identifier.Loc())
		}
		return Params{[]Param{{identifier, nil}}, params.Loc()}
	}
	c.report("Expected typed identifier", params.Loc())
	return Params{}
}
