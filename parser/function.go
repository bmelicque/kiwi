package parser

type FunctionExpression struct {
	TypeParams *Params
	Params     *ParenthesizedExpression // contains *TupleExpression
	Explicit   Expression               // Explicit return type (if any)
	Body       *Block
	returnType ExpressionType
}

func (f *FunctionExpression) Walk(cb func(Node), skip func(Node) bool) {
	if skip(f) {
		return
	}
	cb(f)
	f.Params.Walk(cb, skip)
	if f.Explicit != nil {
		f.Explicit.Walk(cb, skip)
	}
	if f.Body != nil {
		f.Body.Walk(cb, skip)
	}
}

func (f *FunctionExpression) Loc() Loc {
	loc := Loc{Start: f.Params.Loc().Start, End: Position{}}
	if f.Body == nil {
		loc.End = f.Explicit.Loc().End
	} else {
		loc.End = f.Body.Loc().End
	}
	return loc
}

func (f *FunctionExpression) Type() ExpressionType {
	tp := []Generic{}
	if f.TypeParams != nil {
		for i, param := range f.TypeParams.Params {
			tp[i] = Generic{Name: param.Identifier.Token.Text()}
		}
	}
	tuple, _ := f.Params.Type().(Tuple)
	return Function{tp, &tuple, f.returnType}
}

func (f *FunctionExpression) typeCheck(p *Parser) {
	p.pushScope(NewScope(FunctionScope))
	defer p.dropScope()

	if f.TypeParams != nil {
		addTypeParamsToScope(p.scope, *f.TypeParams)
	}

	if f.Params != nil {
		addParamsToScope(p, f.Params.Expr.(*TupleExpression).Elements)
	}
	f.Body.typeCheck(p)
	f.returnType = getFunctionReturnType(p, f.Explicit, f.Body)
}

type FunctionTypeExpression struct {
	TypeParams *Params
	Params     *ParenthesizedExpression // Contains *TupleExpression
	Expr       Expression
}

func (f *FunctionTypeExpression) Walk(cb func(Node), skip func(Node) bool) {
	if skip(f) {
		return
	}
	cb(f)
	f.Params.Walk(cb, skip)
	if f.Expr != nil {
		f.Expr.Walk(cb, skip)
	}
}

func (f *FunctionTypeExpression) Loc() Loc {
	var start, end Position
	if len(f.TypeParams.Params) > 0 {
		start = f.TypeParams.loc.Start
	} else if f.Params != nil {
		start = f.Params.Loc().Start
	} else {
		start = f.Expr.Loc().Start
	}
	if f.Expr != nil {
		end = f.Loc().End
	} else if f.Params != nil {
		end = f.Params.Loc().End
	} else {
		end = f.TypeParams.loc.End
	}
	return Loc{Start: start, End: end}
}
func (f *FunctionTypeExpression) Type() ExpressionType {
	tp := []Generic{}
	if f.TypeParams != nil {
		for _, param := range f.TypeParams.Params {
			tp = append(tp, Generic{Name: param.Identifier.Token.Text()})
		}
	}
	elements := f.Params.Expr.(*TupleExpression).Elements
	p := Tuple{make([]ExpressionType, len(elements))}
	for i, param := range elements {
		t, _ := param.Type().(Type)
		p.elements[i] = t.Value
	}
	var ret ExpressionType = Primitive{UNKNOWN}
	if f.Expr != nil {
		t, ok := f.Expr.Type().(Type)
		if ok {
			ret = t.Value
		}
	}
	return Type{Function{tp, &p, ret}}
}

func (f *FunctionTypeExpression) typeCheck(p *Parser) {
	tuple := f.Params.Expr.(*TupleExpression)
	for i := range tuple.Elements {
		tuple.Elements[i].typeCheck(p)
		if tuple.Elements[i].Type().Kind() != TYPE {
			p.report("Type expected", tuple.Elements[i].Loc())
		}
	}

	if f.Expr != nil && f.Expr.Type().Kind() != TYPE {
		p.report("Type expected", f.Expr.Loc())
	}
}

func (p *Parser) parseFunctionExpression(bracketed *BracketedExpression) Expression {
	p.pushScope(NewScope(FunctionScope))
	defer p.dropScope()

	var typeParams *Params
	if bracketed != nil {
		typeParams = getValidatedTypeParams(p, bracketed)
	}
	paren := p.parseParenthesizedExpression()

	switch p.Peek().Kind() {
	case SlimArrow:
		p.Consume() // ->
		if paren != nil {
			paren.Expr = makeTuple(paren.Expr)
		}
		old := p.allowBraceParsing
		p.allowBraceParsing = false
		expr := p.parseRange()
		p.allowBraceParsing = old
		if p.Peek().Kind() == LeftBrace {
			p.report("No function body expected", p.Peek().Loc())
		}
		return &FunctionTypeExpression{
			TypeParams: typeParams,
			Params:     paren,
			Expr:       expr,
		}
	case FatArrow:
		p.Consume() // =>
		if paren != nil {
			paren.Expr = makeTuple(paren.Expr)
			validateFunctionParams(p, paren)
		}
		var explicit Expression
		if p.Peek().Kind() != LeftBrace {
			old := p.allowBraceParsing
			p.allowBraceParsing = false
			explicit = p.parseRange()
			p.allowBraceParsing = old
		}
		body := p.parseBlock()
		return &FunctionExpression{
			TypeParams: nil,
			Params:     paren,
			Explicit:   explicit,
			Body:       body,
			returnType: nil,
		}
	default:
		return paren
	}
}

func getValidatedTypeParams(p *Parser, bracketed *BracketedExpression) *Params {
	tuple := makeTuple(bracketed.Expr)

	params := make([]Param, len(tuple.Elements))
	for i := range tuple.Elements {
		identifier, ok := tuple.Elements[i].(*Identifier)
		if !ok || !identifier.IsType() {
			p.report("Type identifier expected", tuple.Elements[i].Loc())
		}
		params[i] = Param{Identifier: identifier}
	}

	return &Params{Params: params, loc: bracketed.loc}
}

func validateFunctionParams(p *Parser, node *ParenthesizedExpression) {
	if node.Expr == nil {
		return
	}

	tuple := node.Expr.(*TupleExpression)
	for _, element := range tuple.Elements {
		param, ok := element.(*Param)
		if !ok {
			p.report("Parameter expected: (name Type)", element.Loc())
			continue
		}
		if param.HasColon {
			p.report(
				"No ':' expected between parameter name and type",
				element.Loc(),
			)
		}
	}

	tuple.reportDuplicatedParams(p)
}

func getFunctionReturnType(p *Parser, explicit Expression, body *Block) ExpressionType {
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
func validateFunctionReturns(p *Parser, body *Block) {
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
	if n, ok := node.(*Exit); ok {
		if n.Operator.Kind() == ReturnKeyword {
			*results = append(*results, *n)
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

func addParamsToScope(p *Parser, tuple []Expression) {
	for _, expr := range tuple {
		param, ok := expr.(*Param)
		if !ok {
			continue
		}
		if param.Complement == nil || param.Complement.Type().Kind() != TYPE {
			p.report("Typing expected", param.Loc())
			p.scope.Add(param.Identifier.Text(), param.Loc(), Primitive{UNKNOWN})
		} else {
			typing, _ := param.Complement.Type().(Type)
			p.scope.Add(param.Identifier.Text(), param.Loc(), typing.Value)
		}
	}
}
