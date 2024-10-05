package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type FunctionExpression struct {
	TypeParams Params
	Params     Params
	ReturnType Expression
	Body       Block
}

func (f FunctionExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: f.Params.loc.Start,
		End:   f.Body.loc.End,
	}
}
func (f FunctionExpression) Type() ExpressionType {
	tp := []Generic{}
	for i, param := range f.TypeParams.Params {
		tp[i] = Generic{Name: param.Identifier.Token.Text()}
	}
	var r ExpressionType
	if f.ReturnType != nil {
		if rt, ok := f.ReturnType.Type().(Type); ok {
			r = rt.Value
		} else {
			r = Primitive{UNKNOWN}
		}
	} else {
		r = f.Body.Type()
	}
	return Function{tp, f.Params.Type().(Tuple), r}
}

type FunctionTypeExpression struct {
	TypeParams Params
	Params     []Expression // Identifier | Literal
	Expr       Expression
}

func (f FunctionTypeExpression) Loc() tokenizer.Loc {
	var start tokenizer.Position
	if len(f.TypeParams.Params) > 0 {
		start = f.TypeParams.loc.Start
	} else if len(f.Params) > 0 {
		start = f.Params[0].Loc().Start
	} else {
		start = f.Expr.Loc().Start
	}
	var end tokenizer.Position
	if f.Expr != nil {
		end = f.Loc().End
	} else if len(f.Params) > 0 {
		end = f.Params[len(f.Params)-1].Loc().End
	} else {
		end = f.TypeParams.loc.End
	}
	return tokenizer.Loc{Start: start, End: end}
}
func (f FunctionTypeExpression) Type() ExpressionType {
	tp := []Generic{}
	for i, param := range f.TypeParams.Params {
		tp[i] = Generic{Name: param.Identifier.Token.Text()}
	}
	p := Tuple{make([]ExpressionType, len(f.Params))}
	for i, param := range f.Params {
		t, _ := param.Type().(Type)
		p.elements[i] = t.Value
	}
	return Type{Function{tp, p, f.Expr.Type().(Type).Value}}
}

func (c *Checker) checkFunctionExpression(f parser.FunctionExpression) Expression {
	c.pushScope(NewScope(FunctionScope))
	defer c.dropScope()

	if f.Params == nil {
		c.report("Parameters expected", f.Loc())
	}
	if f.Operator.Kind() == tokenizer.FAT_ARR {
		return checkFunctionExpression(c, f)
	} else {
		return checkFunctionTypeExpression(c, f)
	}
}

// [TypeParam](param Type) => ReturnType { body }
func checkFunctionExpression(c *Checker, f parser.FunctionExpression) FunctionExpression {
	var typeParams Params
	if f.TypeParams != nil {
		typeParams = c.checkTypeParams(*f.TypeParams)
		addTypeParamsToScope(c.scope, typeParams)
	}
	var params Params
	if f.Params != nil && f.Params.Expr != nil {
		params = c.checkParams(*f.Params)
		addParamsToScope(c, params)
	}
	var expr Expression
	if f.Expr != nil {
		expr = c.checkExpression(f.Expr)
		typing, ok := expr.Type().(Type)
		if !ok {
			c.report("Type expected", expr.Loc())
		}
		c.scope.returnType = typing.Value
	}
	body := c.checkBlock(*f.Body)
	// FIXME:
	if c.scope.returnType != nil && !c.scope.returnType.Extends(body.Type()) {
		c.report("Returned type doesn't match expected return type", expr.Loc())
	}
	return FunctionExpression{typeParams, params, expr, body}
}

// [TypeParam](Param) -> ReturnType
func checkFunctionTypeExpression(c *Checker, f parser.FunctionExpression) FunctionTypeExpression {
	var typeParams Params
	if f.TypeParams != nil {
		typeParams = c.checkTypeParams(*f.TypeParams)
		addTypeParamsToScope(c.scope, typeParams)
	}
	params := checkParamsForFunctionType(c, f.Params)
	expr := checkFunctionTypeReturnedType(c, f)
	if f.Body != nil {
		c.report("No body expected", f.Body.Loc())
	}
	return FunctionTypeExpression{typeParams, params, expr}
}
func checkParamsForFunctionType(c *Checker, params *parser.ParenthesizedExpression) []Expression {
	if params == nil || params.Expr == nil {
		return nil
	}

	var checked []Expression
	if tuple, ok := params.Expr.(parser.TupleExpression); ok {
		checked = make([]Expression, len(tuple.Elements))
	} else {
		checked = []Expression{c.checkExpression(params.Expr)}
	}
	for _, el := range checked {
		if el != nil && el.Type() != nil && el.Type().Kind() != TYPE {
			c.report("Type expected", el.Loc())
		}
	}
	return checked
}
func checkFunctionTypeReturnedType(c *Checker, f parser.FunctionExpression) Expression {
	var expr Expression
	if f.Expr != nil {
		expr = c.checkExpression(f.Expr)
	}
	if expr == nil {
		pos := f.Operator.Loc().End
		c.report("Type expected", tokenizer.Loc{Start: pos, End: pos})
	} else if expr.Type().Kind() != TYPE {
		c.report("Type expected", expr.Loc())
	}
	return expr
}

func addParamsToScope(c *Checker, params Params) {
	for _, param := range params.Params {
		if param.Complement == nil {
			c.report("Typing expected", param.Loc())
			c.scope.Add(param.Identifier.Text(), param.Loc(), Primitive{UNKNOWN})
		} else {
			typing, _ := param.Complement.Type().(Type)
			c.scope.Add(param.Identifier.Text(), param.Loc(), typing.Value)
		}
	}
}
