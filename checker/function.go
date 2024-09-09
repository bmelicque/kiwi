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
	tp := []Generic{}
	for i, param := range f.TypeParams.Params {
		tp[i] = Generic{Name: param.Identifier.Token.Text()}
	}
	return Function{tp, f.Params.Type().(Tuple), f.Expr.Type()}
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
	tp := []Generic{}
	for i, param := range f.TypeParams.Params {
		tp[i] = Generic{Name: param.Identifier.Token.Text()}
	}
	return Function{tp, f.Params.Type().(Tuple), f.ReturnType.Type().(Type).Value}
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
	tp := []Generic{}
	for i, param := range g.TypeParams.Params {
		tp[i] = Generic{Name: param.Identifier.Token.Text()}
	}
	return Function{tp, Tuple{}, g.Expr.Type()}
}

func (c *Checker) checkFunctionExpression(f parser.FunctionExpression) Expression {
	c.pushScope(NewScope())
	defer c.dropScope()

	typeParams := c.handleFunctionTypeParams(f.TypeParams)
	params := c.handleFunctionParams(f.Params)
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

func (c *Checker) handleFunctionTypeParams(expr *parser.BracketedExpression) Params {
	if expr == nil {
		return Params{}
	}
	if expr.Expr == nil {
		return Params{[]Param{}, expr.Loc()}
	}
	params := c.checkParams(expr.Expr)
	for _, param := range params.Params {
		if param.Typing == nil {
			name := param.Identifier.Text()
			t := Type{TypeAlias{Name: name, Ref: Generic{Name: name}}}
			c.scope.Add(name, param.Loc(), t)
		} else {
			// TODO: constrained generic
		}
	}
	return params
}

func (c *Checker) handleFunctionParams(expr *parser.ParenthesizedExpression) Params {
	if expr == nil {
		return Params{}
	}
	if expr.Expr == nil {
		return Params{[]Param{}, expr.Loc()}
	}
	params := c.checkParams(expr.Expr)
	for _, param := range params.Params {
		if param.Typing == nil {
			c.report("Typing expected", param.Loc())
			c.scope.Add(param.Identifier.Text(), param.Loc(), Primitive{UNKNOWN})
		} else {
			typing, _ := param.Typing.Type().(Type)
			c.scope.Add(param.Identifier.Text(), param.Loc(), typing.Value)
		}
	}
	return params
}
