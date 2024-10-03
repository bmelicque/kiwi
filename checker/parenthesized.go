package checker

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

// (Expression)
type ParenthesizedExpression struct {
	Expr Expression
	loc  tokenizer.Loc
}

func (p ParenthesizedExpression) Loc() tokenizer.Loc { return p.loc }
func (p ParenthesizedExpression) Type() ExpressionType {
	if p.Expr == nil {
		return Primitive{NIL}
	}
	if param, ok := p.Expr.(Param); ok {
		return Type{Object{map[string]ExpressionType{
			param.Identifier.Text(): param.Complement.Type(),
		}}}
	}
	return p.Expr.Type()
}

func (c *Checker) checkParenthesizedExpression(node parser.ParenthesizedExpression) ParenthesizedExpression {
	if node.Expr == nil {
		return ParenthesizedExpression{loc: node.Loc()}
	}
	return ParenthesizedExpression{
		Expr: c.checkExpression(node.Expr),
		loc:  node.Loc(),
	}
}

// (param ParamType, otherParam OtherParamType)
type Params struct {
	Params []Param
	loc    tokenizer.Loc
}

func (p Params) Loc() tokenizer.Loc { return p.loc }

// FIXME: object
func (p Params) Type() ExpressionType {
	types := make([]ExpressionType, len(p.Params))
	for i, element := range p.Params {
		types[i] = element.Type()
	}
	return Tuple{types}
}

// (identifier Type)
func (c *Checker) checkParams(node parser.ParenthesizedExpression) Params {
	return checkParamList(c, node.Expr, c.checkParam)
}

// (Identifier Type?)
func (c *Checker) checkTypeParams(node parser.BracketedExpression) Params {
	return checkParamList(c, node.Expr, c.checkTypeParam)
}

// (identifier: value)
func (c *Checker) checkNamedArguments(node parser.ParenthesizedExpression) Params {
	return checkParamList(c, node.Expr, c.checkNamedArgument)
}

func (c *Checker) checkArguments(node parser.ParenthesizedExpression) Params {
	return checkParamList(c, node.Expr, c.checkArgument)
}

func checkParamList(c *Checker, node parser.Node, checkSingle func(parser.Node) Param) Params {
	if node == nil {
		// FIXME: loc
		return Params{}
	}
	params := Params{loc: node.Loc()}
	tuple, ok := node.(parser.TupleExpression)
	if !ok {
		params.Params = []Param{checkSingle(node)}
		return params
	}
	params.Params = make([]Param, len(tuple.Elements))
	for i, element := range tuple.Elements {
		params.Params[i] = checkSingle(element)
	}
	checkParamDuplicates(c, params.Params)
	return params
}
func checkParamDuplicates(c *Checker, params []Param) {
	declarations := map[string][]tokenizer.Loc{}
	for _, param := range params {
		name := param.Identifier.Text()
		if name != "" {
			declarations[name] = append(declarations[name], param.Identifier.Loc())
		}
	}
	for name, locs := range declarations {
		if len(locs) == 1 {
			continue
		}
		for _, loc := range locs {
			c.report(fmt.Sprintf("Duplicate identifier '%v'", name), loc)
		}
	}
}

// param ParamType
type Param struct {
	Identifier Identifier
	Complement Expression // Type for params, value for arguments
	loc        tokenizer.Loc
}

func (p Param) Loc() tokenizer.Loc { return p.loc }
func (p Param) Type() ExpressionType {
	if p.Complement == nil {
		return Primitive{UNKNOWN}
	}
	typing, ok := p.Complement.Type().(Type)
	if !ok {
		return Primitive{UNKNOWN}
	}
	return typing.Value
}

// identifier Type
func (c *Checker) checkParam(node parser.Node) Param {
	expr, ok := node.(parser.TypedExpression)
	if !ok {
		c.report("Identifier and Type expected", node.Loc())
		return Param{}
	}

	identifier := checkParamIdentifier(c, expr.Expr)
	typing := checkParamTyping(c, expr.Typing)

	if expr.Colon {
		c.report("Expected type (no use of ':')", expr.Typing.Loc())
	}

	return Param{identifier, typing, node.Loc()}
}
func checkParamIdentifier(c *Checker, node parser.Node) Identifier {
	var identifier Identifier
	if token, ok := node.(parser.TokenExpression); ok {
		identifier, _ = c.checkToken(token, false).(Identifier)
	}
	if identifier == (Identifier{}) {
		c.report("Identifier expected", node.Loc())
	}
	return identifier
}
func checkParamTyping(c *Checker, node parser.Node) Expression {
	if node == nil {
		return nil
	}
	typing := c.checkExpression(node)
	if _, ok := typing.Type().(Type); !ok {
		c.report("Typing expected", node.Loc())
	}
	return typing
}

func (c *Checker) checkTypeParam(node parser.Node) Param {
	var param Param
	if _, ok := node.(parser.TypedExpression); ok {
		param = c.checkParam(node)
	} else {
		param = Param{checkParamIdentifier(c, node), nil, node.Loc()}
	}
	if !param.Identifier.isType {
		c.report("Type name expected", param.Identifier.Loc())
	}
	return param
}
func (c *Checker) checkNamedArgument(node parser.Node) Param {
	expr, ok := node.(parser.TypedExpression)
	if !ok {
		c.report("Identifier and value expected", node.Loc())
		return Param{}
	}

	identifier := checkParamIdentifier(c, expr.Expr)
	value := c.checkExpression(expr.Typing)

	if !expr.Colon {
		c.report("':' expected", expr.Typing.Loc())
	}

	return Param{identifier, value, node.Loc()}
}
func (c *Checker) checkArgument(node parser.Node) Param {
	expr, ok := node.(parser.TypedExpression)
	if ok {
		c.report("Expression expected", node.Loc())
		return Param{Complement: c.checkExpression(expr.Expr)}
	}
	value := c.checkExpression(node)
	return Param{Complement: value, loc: node.Loc()}
}
