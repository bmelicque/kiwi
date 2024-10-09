package parser

type FunctionExpression struct {
	TypeParams *Params
	Params     *Params
	Explicit   Expression // Explicit return type (if any)
	Body       *Block
	returnType ExpressionType
}

func (f FunctionExpression) Loc() Loc {
	loc := Loc{Start: f.Params.Loc().Start, End: Position{}}
	if f.Body == nil {
		loc.End = f.Explicit.Loc().End
	} else {
		loc.End = f.Body.Loc().End
	}
	return loc
}

func (f FunctionExpression) Type() ExpressionType {
	tp := []Generic{}
	for i, param := range f.TypeParams.Params {
		tp[i] = Generic{Name: param.Identifier.Token.Text()}
	}
	tuple, _ := f.Params.Type().(Tuple)
	return Function{tp, &tuple, f.returnType}
}

type FunctionTypeExpression struct {
	TypeParams *Params
	Params     *Params
	Expr       Expression
}

func (f FunctionTypeExpression) Loc() Loc {
	var start, end Position
	if len(f.TypeParams.Params) > 0 {
		start = f.TypeParams.loc.Start
	} else if len(f.Params.Params) > 0 {
		start = f.Params.Params[0].Loc().Start
	} else {
		start = f.Expr.Loc().Start
	}
	if f.Expr != nil {
		end = f.Loc().End
	} else if len(f.Params.Params) > 0 {
		end = f.Params.Params[len(f.Params.Params)-1].Loc().End
	} else {
		end = f.TypeParams.loc.End
	}
	return Loc{Start: start, End: end}
}
func (f FunctionTypeExpression) Type() ExpressionType {
	tp := []Generic{}
	for i, param := range f.TypeParams.Params {
		tp[i] = Generic{Name: param.Identifier.Token.Text()}
	}
	p := Tuple{make([]ExpressionType, len(f.Params.Params))}
	for i, param := range f.Params.Params {
		t, _ := param.Type().(Type)
		p.elements[i] = t.Value
	}
	return Type{Function{tp, &p, f.Expr.Type().(Type).Value}}
}

func (p *Parser) parseFunctionExpression(brackets *BracketedExpression) Expression {
	p.pushScope(NewScope(FunctionScope))
	defer p.dropScope()

	var typeParams *Params
	if brackets != nil {
		typeParams = p.getValidatedTypeParams(*brackets)
		addTypeParamsToScope(p.scope, *typeParams)
	}
	// Important: make sure that potential type params are already added to scope
	paren := p.parseParenthesizedExpression()

	switch p.Peek().Kind() {
	case SlimArrow:
		p.Consume() // ->
		params := p.getValidatedTypeList(*paren)
		old := p.allowBraceParsing
		p.allowBraceParsing = false
		expr := p.parseRange()
		p.allowBraceParsing = old
		// expr == nil already reported while parsing
		if expr != nil && expr.Type().Kind() != TYPE {
			p.report("Type expected", expr.Loc())
		}
		if p.Peek().Kind() == LeftBrace {
			p.report("No function body expected", p.Peek().Loc())
		}
		return &FunctionTypeExpression{
			TypeParams: typeParams,
			Params:     params,
			Expr:       expr,
		}
	case FatArrow:
		p.Consume() // =>
		params := p.getValidatedParams(*paren)
		if params != nil {
			addParamsToScope(p, *params)
		}
		var explicit Expression
		if p.Peek().Kind() != LeftBrace {
			old := p.allowBraceParsing
			p.allowBraceParsing = false
			explicit = p.parseRange()
			p.allowBraceParsing = old
		}
		body := p.parseBlock()
		returnedType := getFunctionReturnType(p, explicit, *body)
		return &FunctionExpression{
			TypeParams: nil,
			Params:     params,
			Explicit:   explicit,
			Body:       body,
			returnType: returnedType,
		}
	default:
		return paren
	}
}

func getFunctionReturnType(p *Parser, explicit Expression, body Block) ExpressionType {
	validateFunctionReturns(p, body)
	if explicit == nil {
		return body.Type()
	}
	t, ok := explicit.Type().(Type)
	if !ok {
		p.report("Type expected", explicit.Loc())
		return Primitive{UNKNOWN}
	}
	if !t.Value.Extends(body.Type()) {
		p.report("Returned type doesn't match expected return type", body.reportLoc())
	}
	return t.Value
}

// Check if every return statement inside a body matches the body's type
func validateFunctionReturns(p *Parser, body Block) {
	returns := []Exit{}
	findReturnStatements(body, &returns)
	bType := body.Type()
	ok := true
	for _, r := range returns {
		var t ExpressionType
		if r.Value != nil {
			t = r.Value.Type()
		} else {
			t = Primitive{NIL}
		}
		if !bType.Extends(t) {
			ok = false
			p.report("Mismatched types", r.Value.Loc())
		}
	}
	if !ok {
		p.report("Mismatched types", body.reportLoc())
	}
}
func findReturnStatements(node Node, results *[]Exit) {
	if node == nil {
		return
	}
	if n, ok := node.(Exit); ok {
		if n.Operator.Kind() == ReturnKeyword {
			*results = append(*results, n)
		}
		return
	}
	switch node := node.(type) {
	case *Block:
		for _, statement := range node.Statements {
			findReturnStatements(statement, results)
		}
	case *IfExpression:
		findReturnStatements(node.Body, results)
		findReturnStatements(node.Alternate, results)
	}
}

func addParamsToScope(p *Parser, params Params) {
	for _, param := range params.Params {
		if param.Complement == nil {
			p.report("Typing expected", param.Loc())
			p.scope.Add(param.Identifier.Text(), param.Loc(), Primitive{UNKNOWN})
		} else {
			typing, _ := param.Complement.Type().(Type)
			p.scope.Add(param.Identifier.Text(), param.Loc(), typing.Value)
		}
	}
}
