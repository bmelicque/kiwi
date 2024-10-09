package parser

import "fmt"

// (param ParamType, otherParam OtherParamType)
type Params struct {
	Params []Param
	loc    Loc
}

func (p Params) Loc() Loc { return p.loc }

// FIXME: object
func (p Params) Type() ExpressionType {
	types := make([]ExpressionType, len(p.Params))
	for i, element := range p.Params {
		types[i] = element.Type()
	}
	return Tuple{types}
}

// Take an expression grouped between parentheses and tries to parse it as params
// The expected form is:
// (identifier Type, identifier Type, ...)
func (p *Parser) getValidatedParams(node ParenthesizedExpression) *Params {
	return getValidatedParamList(p, node.Expr, p.getValidatedParam)
}

// Take an expression grouped between brackets and tries to parse it as type params
// The expected form is:
// [TypeIdentifier, TypeIdentifier TypeConstraint, ...]
func (p *Parser) getValidatedTypeParams(node BracketedExpression) *Params {
	return getValidatedParamList(p, node.Expr, p.getValidatedTypeParam)
}

// Take an expression grouped between parentheses and tries to parse it as named arguments
// The expected form is:
// (identifier: value, identifier: value, ...)
func (p *Parser) getValidatedNamedArguments(node ParenthesizedExpression) *Params {
	return getValidatedParamList(p, node.Expr, p.getValidatedNamedArgument)
}

// Take an expression grouped between parentheses and tries to parse it as arguments
// The expected form is:
// (value, value, ...)
func (p *Parser) getValidatedArguments(node ParenthesizedExpression) *Params {
	return getValidatedParamList(p, node.Expr, p.getValidatedArgument)
}

// Take an expression grouped between parentheses and tries to parse it as type list
// The expected form is:
// (Type, Type, ...)
// This is useful mostly for parsing function types
func (p *Parser) getValidatedTypeList(node ParenthesizedExpression) *Params {
	return getValidatedParamList(p, node.Expr, p.getValidatedType)
}

func getValidatedParamList(p *Parser, node Node, validateOne func(Node) Param) *Params {
	if node == nil {
		// FIXME: loc
		return nil
	}
	params := &Params{loc: node.Loc()}
	tuple, ok := node.(*TupleExpression)
	if !ok {
		params.Params = []Param{validateOne(node)}
		return params
	}
	params.Params = make([]Param, len(tuple.Elements))
	for i, element := range tuple.Elements {
		params.Params[i] = validateOne(element)
	}
	reportDuplicatedParams(p, params.Params)
	return params
}
func reportDuplicatedParams(p *Parser, params []Param) {
	declarations := map[string][]Loc{}
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
			p.report(fmt.Sprintf("Duplicate identifier '%v'", name), loc)
		}
	}
}

// An identifier followed by a type expression
type Param struct {
	Identifier *Identifier
	Complement Expression // Type for params, value for arguments
}

func (p Param) Loc() Loc {
	var start, end Position
	if p.Identifier != nil {
		start = p.Identifier.Loc().Start
	} else {
		start = p.Complement.Loc().Start
	}
	if p.Complement != nil {
		end = p.Complement.Loc().End
	} else {
		end = p.Identifier.Loc().End
	}
	return Loc{start, end}
}
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
func (p *Parser) getValidatedParam(node Node) Param {
	expr, ok := node.(TypedExpression)
	if !ok {
		p.report("Identifier and Type expected", node.Loc())
		return Param{}
	}

	identifier := checkParamIdentifier(p, expr.Expr)
	validateParamTyping(p, expr.Typing)

	if expr.Colon {
		p.report("Expected type (no use of ':')", expr.Typing.Loc())
	}

	return Param{identifier, expr.Typing}
}

func (p *Parser) getValidatedTypeParam(node Node) Param {
	var param Param
	if _, ok := node.(TypedExpression); ok {
		param = p.getValidatedParam(node)
	} else {
		param = Param{checkParamIdentifier(p, node), nil}
	}
	if !param.Identifier.isType {
		p.report("Type name expected", param.Identifier.Loc())
	}
	return param
}

// Take a node and try to parse it as a named argument.
// The expected form is "name: value"
func (p *Parser) getValidatedNamedArgument(node Node) Param {
	expr, ok := node.(TypedExpression)
	if !ok {
		p.report("Identifier and value expected", node.Loc())
		return Param{}
	}

	identifier := checkParamIdentifier(p, expr.Expr)

	if !expr.Colon {
		p.report("':' expected", expr.Typing.Loc())
	}

	return Param{identifier, expr.Typing}
}

// Take an expression and makes sure it is a valid argument for a function
func (p *Parser) getValidatedArgument(node Node) Param {
	expr, ok := node.(TypedExpression)
	if ok {
		p.report("Missing comma between expressions", node.Loc())
		return Param{Complement: expr.Typing}
	}
	value, ok := node.(Expression)
	if !ok {
		p.report("Expression expected", node.Loc())
	}
	return Param{Complement: value}
}

// Take an expression and makes sure it's a valid type (for function types)
func (p *Parser) getValidatedType(node Node) Param {
	param := p.getValidatedArgument(node)
	if param.Complement.Type().Kind() != TYPE {
		p.report("Type expected", param.Complement.Loc())
	}
	return param
}

func checkParamIdentifier(p *Parser, node Node) *Identifier {
	identifier, ok := node.(*Identifier)
	if !ok {
		p.report("Identifier expected", node.Loc())
	}
	return identifier
}
func validateParamTyping(p *Parser, expr Expression) {
	if expr == nil {
		return
	}
	if _, ok := expr.Type().(Type); !ok {
		p.report("Typing expected", expr.Loc())
	}
}
