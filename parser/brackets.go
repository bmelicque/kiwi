package parser

// Expression between brackets, such as `[Type]`
type BracketedExpression struct {
	Expr Expression
	loc  Loc
}

func (b *BracketedExpression) getChildren() []Node {
	if b.Expr == nil {
		return []Node{}
	}
	return []Node{b.Expr}
}

func (b *BracketedExpression) typeCheck(p *Parser) {
	b.Expr.typeCheck(p)
}

func (b *BracketedExpression) Loc() Loc             { return b.loc }
func (b *BracketedExpression) Type() ExpressionType { return nil }
func (p *Parser) parseBracketedExpression() *BracketedExpression {
	if p.Peek().Kind() != LeftBracket {
		panic("'[' expected!")
	}
	loc := p.Consume().Loc()

	outer := p.allowEmptyExpr
	p.allowEmptyExpr = true
	expr := ParseExpression(p)
	p.allowEmptyExpr = outer

	next := p.Peek()
	if next.Kind() != RightBracket {
		p.report("']' expected", next.Loc())
		if expr != nil {
			loc.End = expr.Loc().End
		}
	} else {
		loc.End = p.Consume().Loc().End
	}

	return &BracketedExpression{expr, loc}
}

// Validate type params, either for generic type or function expression.
// Turn Expr into a *TupleExpression containing *Param elements.
func validateTypeParams(p *Parser, bracketed *BracketedExpression) {
	if bracketed.Expr == nil {
		p.report("Type params expected between brackets", bracketed.loc)
		return
	}

	tuple := makeTuple(bracketed.Expr)
	for i := range tuple.Elements {
		tuple.Elements[i] = getValidatedTypeParam(p, tuple.Elements[i])
	}

	tuple.reportDuplicatedParams(p)
	bracketed.Expr = tuple
}

func getValidatedTypeParam(p *Parser, expr Expression) *Param {
	param, ok := expr.(*Param)
	if !ok {
		identifier, ok := expr.(*Identifier)
		if !ok || !identifier.IsType() {
			p.report("Type identifier expected", expr.Loc())
			return &Param{}
		}
		return &Param{Identifier: identifier}
	}

	if param.HasColon {
		p.report(
			"No ':' expected between identifier and constraint",
			param.Loc(),
		)
	}
	return param
}

func typeCheckTypeParams(p *Parser, typeParams *BracketedExpression) {
	if typeParams == nil {
		return
	}

	tuple := typeParams.Expr.(*TupleExpression)
	for i := range tuple.Elements {
		param := tuple.Elements[i].(*Param)
		if param.Complement == nil {
			continue
		}
		param.Complement.typeCheck(p)
		if _, ok := param.Complement.Type().(Type); !ok {
			p.report("Type expected", param.Complement.Loc())
		}
	}
	addTypeParamsToScope(p.scope, typeParams)
}
