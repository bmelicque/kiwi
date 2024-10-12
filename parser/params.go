package parser

import "fmt"

// (param ParamType, otherParam OtherParamType)
type Params struct {
	Params []Param
	raw    *TupleExpression
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

func (params *Params) typeCheck(p *Parser) {
	for i := range params.Params {
		params.Params[i].Complement.typeCheck(p)
	}
}

// next token should be a '('
func (p *Parser) parseArguments() *Params {
	token := p.Consume() //
	start := token.Loc().Start

	expr := makeTuple(p.parseTupleExpression())
	params := make([]Param, len(expr.Elements))
	for i := range expr.Elements {
		params[i] = getValidatedArgument(p, expr.Elements[i])
	}

	if p.Peek().Kind() != RightParenthesis {
		var end Position
		if len(expr.Elements) == 0 {
			end = token.Loc().End
		} else {
			end = expr.Loc().End
		}
		p.report("')' expected", p.Peek().Loc())
		return &Params{Params: params, loc: Loc{start, end}}
	}
	end := p.Consume().Loc().End
	return &Params{Params: params, loc: Loc{start, end}}
}

// next token should be '(' or '['
func (p *Parser) parseParamsRaw() *Params {
	token := p.Consume() // '(' or '['
	start := token.Loc().Start

	expr := makeTuple(p.parseTupleExpression())

	if token.Kind() == LeftParenthesis && p.Peek().Kind() != RightParenthesis {
		recover(p, RightParenthesis)
	}
	if token.Kind() == LeftBracket && p.Peek().Kind() != RightBracket {
		recover(p, RightBracket)
	}
	end := p.Peek().Loc().End
	if p.Peek().Kind() == RightParenthesis || p.Peek().Kind() == RightBracket {
		p.Consume()
	}
	return &Params{raw: expr, loc: Loc{start, end}}
}

type paramValidator = func(*Parser, Expression) Param

func (params *Params) validate(p *Parser, validator paramValidator) {
	params.Params = make([]Param, len(params.raw.Elements))
	for i := range params.Params {
		params.Params[i] = validator(p, params.raw.Elements[i])
	}
	reportDuplicatedParams(p, params.Params)
	// TODO: check that param kinds match (no mixing args and names args)
	// return raw to GC
	params.raw = nil
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

type ParamKind = int8

const (
	RegularParam ParamKind = iota
	Argument
	NamedArgument
)

// An identifier followed by a type expression
type Param struct {
	Identifier *Identifier
	Complement Expression // Type for params, value for arguments
	HasColon   bool
	kind
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

func (p *Parser) getValidatedParam(node Expression) Param {
	expr, ok := node.(*TypedExpression)
	if !ok {
		return Param{Complement: node}
	}

	identifier, ok := expr.Expr.(*Identifier)
	if !ok {
		p.report("Identifier expected", node.Loc())
	}

	if expr.Colon {
		p.report("No ':' expected between name and type", node.Loc())
	}
	return Param{Identifier: identifier, Complement: expr.Typing}
}

func (p *Parser) getValidatedTypeParam(node Expression) Param {
	expr, ok := node.(*TypedExpression)
	if !ok {
		identifier, ok := node.(*Identifier)
		if !ok || !identifier.isType {
			p.report("Type identifier expected", node.Loc())
		}
		return Param{Identifier: identifier}
	}

	identifier, ok := expr.Expr.(*Identifier)
	if !ok || !identifier.isType {
		p.report("Type identifier expected", expr.Expr.Loc())
	}

	if expr.Colon {
		p.report("No ':' expected between name and type", node.Loc())
	}
	return Param{Identifier: identifier, Complement: expr.Typing}
}

func getValidatedNamedArgument(p *Parser, node Expression) Param {
	expr, ok := node.(*TypedExpression)
	if !ok {
		p.report("Named argument expected", node.Loc())
		return Param{Complement: node}
	}

	identifier, ok := expr.Expr.(*Identifier)
	if !ok {
		p.report("Name expected", expr.Expr.Loc())
	}

	if expr.Colon {
		p.report("':' expected between name and value", node.Loc())
	}
	return Param{identifier, expr.Typing, true}
}

func getValidatedArgument(p *Parser, node Expression) Param {
	typed, ok := node.(*TypedExpression)
	if !ok {
		return Param{Complement: node, kind: Argument}
	}
	identifier, ok := typed.Expr.(*Identifier)
	if !ok {
		p.report("Name expected", node.Loc())
	}
	if expr.Colon {
		p.report("':' expected between name and value", node.Loc())
	}
	return Param{
		Identifier: identifier,
		Complement: typed.Typing,
		kind:       NamedArgument,
	}
}
