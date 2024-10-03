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
	Body       Block
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
	c.pushScope(NewScope())
	defer c.dropScope()

	if f.TypeParams != nil && f.Params == nil {
		c.report("Parameters expected", f.Loc())
	}

	isType, ok := isFunctionType(f.Params)
	if !ok && f.Operator.Kind() == tokenizer.FAT_ARR {
		return checkFatArrowFunction(c, f)
	}
	if !ok {
		return checkUnknownFunctionExpression(c, f)
	}
	if isType {
		return checkFunctionTypeExpression(c, f)
	}

	if f.Operator.Kind() == tokenizer.FAT_ARR {
		return checkFatArrowFunction(c, f)
	}
	return checkSlimArrowFunction(c, f)
}

// [TypeParam](param Type) => ReturnType { body }
func checkFatArrowFunction(c *Checker, f parser.FunctionExpression) FatArrowFunction {
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
	return FatArrowFunction{typeParams, params, expr, body}
}

// [TypeParam](param Type) -> returnValue
func checkSlimArrowFunction(c *Checker, f parser.FunctionExpression) SlimArrowFunction {
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
		if expr.Type().Kind() == TYPE {
			c.report("Value expected, got type", expr.Loc())
		}
	}
	if f.Body != nil {
		c.report("No body expected", f.Body.Loc())
	}
	return SlimArrowFunction{typeParams, params, expr}
}

// [TypeParam](Param) -> ReturnType
func checkFunctionTypeExpression(c *Checker, f parser.FunctionExpression) FunctionTypeExpression {
	var typeParams Params
	if f.TypeParams != nil {
		typeParams = c.checkTypeParams(*f.TypeParams)
		addTypeParamsToScope(c.scope, typeParams)
	}
	params := checkParamsForFunctionType(c, f.Params)
	if f.Operator.Kind() == tokenizer.FAT_ARR {
		c.report("'->' expected", f.Operator.Loc())
	}
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

// [TypeParam]() -> something
func checkUnknownFunctionExpression(c *Checker, f parser.FunctionExpression) Expression {
	var typeParams Params
	if f.TypeParams != nil {
		typeParams = c.checkTypeParams(*f.TypeParams)
		addTypeParamsToScope(c.scope, typeParams)
	}
	if f.Body != nil {
		c.report("No body expected", f.Body.Loc())
	}
	if f.Expr == nil {
		return SlimArrowFunction{typeParams, Params{}, nil}
	}

	expr := c.checkExpression(f.Expr)
	if expr.Type().Kind() == TYPE {
		return FunctionTypeExpression{typeParams, []Expression{}, expr}
	}
	return SlimArrowFunction{typeParams, Params{}, expr}
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

// (isType, isDefined)
func isFunctionType(params *parser.ParenthesizedExpression) (bool, bool) {
	if params == nil || params.Expr == nil {
		return false, false
	}

	var first parser.Node
	if tuple, ok := params.Expr.(parser.TupleExpression); ok {
		first = tuple.Elements[0]
	} else {
		first = params.Expr
	}

	if _, ok := first.(parser.TypedExpression); ok {
		return false, true
	}
	return true, true
}
