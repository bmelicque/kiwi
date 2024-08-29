package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type SlimArrowFunction struct {
	TypeParams Params
	Params     Params
	Expr       Expression
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
	TypeParams Params
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

type GenericTypeDef struct {
	TypeParams Params
	Expr       Expression
}

func (g GenericTypeDef) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: g.TypeParams.loc.Start,
		End:   g.Expr.Loc().End,
	}
}
func (g GenericTypeDef) Type() ExpressionType {
	return Function{g.TypeParams.Type(), g.Expr.Type()}
}

func (c *Checker) checkFunctionExpression(f parser.FunctionExpression) Expression {
	c.pushScope(NewScope())
	defer c.dropScope()

	typeParams := Params{}
	if f.TypeParams != nil {
		typeParams = c.handleFunctionParams(f.TypeParams.Expr, f.TypeParams.Loc())
	}
	params := Params{}
	if f.Params != nil {
		params = c.handleFunctionParams(f.Params.Expr, f.Params.Loc())
	}
	expr := c.checkExpression(f.Expr)

	switch f.Operator.Kind() {
	case tokenizer.SLIM_ARR:
		if f.Body != nil {
			pos := f.Expr.Loc().End
			c.report("Expected no body", tokenizer.Loc{Start: pos, End: pos})
		}
		if f.TypeParams != nil && f.Params == nil {
			_, returnsType := expr.Type().(Type)
			if !returnsType {
				c.report("Expected type", f.Expr.Loc())
			}
			return GenericTypeDef{typeParams, expr}
		}
		return SlimArrowFunction{typeParams, params, expr}
	case tokenizer.FAT_ARR:
		typing, ok := expr.Type().(Type)
		if !ok {
			c.report("Expected type", f.Expr.Loc())
		}
		c.scope.returnType = typing.Value
		body := c.checkBody(*f.Body)
		return FatArrowFunction{typeParams, params, expr, body}
	default:
		panic("Unexpected token while checking function expression")
	}
}

func (c *Checker) handleFunctionParams(expr parser.Node, defaultLoc tokenizer.Loc) Params {
	params := Params{[]Param{}, defaultLoc}
	if expr != nil {
		params = c.checkParams(expr)
	}
	for _, param := range params.Params {
		var typing Type
		if param.Typing == nil {
			c.scope.Add(param.Identifier.Text(), param.Loc(), Type{Primitive{UNKNOWN}})
		} else {
			typing, _ = param.Typing.Type().(Type)
			c.scope.Add(param.Identifier.Text(), param.Loc(), typing.Value)
		}
	}
	return params
}
