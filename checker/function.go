package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type SlimArrowFunction struct {
	Params Params
	Expr   Expression
}

func (f SlimArrowFunction) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: f.Params.loc.Start,
		End:   f.Expr.Loc().End,
	}
}
func (f SlimArrowFunction) Type() ExpressionType {
	return Function{f.Params.Type(), f.Expr.Type()}
}

type FatArrowFunction struct {
	Params     Params
	ReturnType Expression
	Body       Body
}

func (f FatArrowFunction) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: f.Params.loc.Start,
		End:   f.Body.loc.End,
	}
}
func (f FatArrowFunction) Type() ExpressionType {
	return Function{f.Params.Type(), f.ReturnType.Type().(Type).Value}
}

func (c *Checker) checkFunctionExpression(f parser.FunctionExpression) Node {
	c.pushScope(NewScope())
	defer c.dropScope()

	params := c.checkParams(f.Params)
	for _, param := range params.Params {
		typing, ok := param.Typing.Type().(Type)
		if ok {
			c.scope.Add(param.Identifier.Text(), param.Loc(), typing.Value)
		}
	}

	expr := c.CheckExpression(f.Expr)

	switch f.Operator.Kind() {
	case tokenizer.SLIM_ARR:
		if f.Body != nil {
			pos := f.Expr.Loc().End
			c.report("Expected no body", tokenizer.Loc{Start: pos, End: pos})
		}
		return SlimArrowFunction{params, expr}
	case tokenizer.FAT_ARR:
		typing, ok := expr.Type().(Type)
		if !ok {
			c.report("Expected type", f.Expr.Loc())
		}
		c.scope.returnType = typing.Value
		body := c.checkBody(*f.Body)
		return FatArrowFunction{params, expr, body}
	default:
		panic("Unexpected token while checking function expression")
	}
}
