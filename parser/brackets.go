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
	if b.Expr != nil {
		b.Expr.typeCheck(p)
	}
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
	expr := p.parseRange()
	p.allowEmptyExpr = outer

	next := p.Peek()
	if next.Kind() != RightBracket {
		p.error(&Literal{next}, RightBracketExpected)
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
		p.error(bracketed, TypeParamsExpected)
		return
	}

	tuple := MakeTuple(bracketed.Expr)
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
			p.error(expr, TypeIdentifierExpected)
			return &Param{}
		}
		return &Param{Identifier: identifier}
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
			p.error(param.Complement, TypeExpected)
		}
	}
	addTypeParamsToScope(p.scope, typeParams)
}

// Type-check given type arguments against given expected type params.
func typeCheckTypeArgs(p *Parser, args *TupleExpression, expected []Generic) {
	var l int
	if args != nil {
		args.typeCheck(p)
		l = len(args.Elements)
	}

	if l > len(expected) {
		p.error(args, TooManyElements, len(expected), len(args.Elements))
	}

	for i := range expected {
		if i < l {
			typeCheckTypeArg(p, args.Elements[i], &expected[i])
		} else {
			addGenericToScope(p.scope, expected[i], Loc{})
		}
	}
}

func typeCheckTypeArg(p *Parser, arg Expression, expected *Generic) {
	typing, ok := arg.Type().(Type)
	if !ok {
		p.error(arg, TypeExpected)
		return
	}
	if expected.Value == nil {
		if expected.Constraints != nil && !expected.Constraints.Extends(typing) {
			p.error(arg, CannotAssignType, expected.Constraints, typing)
		} else {
			(*expected).Value = typing.Value
		}
	} else if !expected.Value.Extends(typing) {
		p.error(arg, CannotAssignType, expected.Value, typing)
	}
	addGenericToScope(p.scope, *expected, arg.Loc())
}

func addGenericToScope(scope *Scope, generic Generic, loc Loc) {
	scope.Add(generic.Name, loc, Type{generic})
	v, _ := scope.Find(generic.Name)
	v.readAt(loc)
}

func (b *BracketedExpression) getGenerics() []Generic {
	elements := b.Expr.(*TupleExpression).Elements
	generics := make([]Generic, len(elements))
	for i := range elements {
		param := elements[i].(*Param)
		generics[i] = Generic{Name: param.Identifier.Text()}
		if param.Complement == nil {
			continue
		}
		if t, ok := param.Complement.Type().(Type); ok {
			generics[i].Constraints = t.Value
		}
	}
	return generics
}
