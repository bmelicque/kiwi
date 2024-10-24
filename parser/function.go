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

	if f.Params != nil && f.Params.Expr != nil {
		addParamsToScope(p, f.Params.Expr.(*TupleExpression).Elements)
	}
	f.Body.typeCheck(p)

	if f.Explicit != nil {
		typeCheckExplicitReturn(p, f)
	} else {
		typeCheckImplicitReturn(p, f)
	}
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

// Validate all of a function params' structures
func validateFunctionParams(p *Parser, node *ParenthesizedExpression) {
	if node.Expr == nil {
		return
	}

	tuple := node.Expr.(*TupleExpression)
	for _, element := range tuple.Elements {
		validateFunctionParam(p, element)
	}

	tuple.reportDuplicatedParams(p)
}

// Validate a function param's structure
func validateFunctionParam(p *Parser, expr Expression) {
	param, ok := expr.(*Param)
	if !ok {
		p.report("Parameter expected: (name Type)", expr.Loc())
		return
	}
	if param.HasColon {
		p.report("No ':' expected between parameter name and type", expr.Loc())
	}
}

// Type check all possible return points against the explicit return type.
// Also check possible failure points.
func typeCheckExplicitReturn(p *Parser, f *FunctionExpression) {
	explicit := f.Explicit.Type()
	if t, ok := explicit.(Type); ok {
		explicit = t.Value
	} else {
		explicit = Primitive{UNKNOWN}
	}
	f.returnType = explicit

	typeCheckHappyReturn(p, f.Body, explicit)

	err := getErrorType(explicit)
	tries := findTryExpressions(f.Body)
	for _, t := range tries {
		if !err.Extends(getErrorType(t.Expr.Type())) {
			p.report("Error type doesn't match expected type", t.Expr.Loc())
		}
	}
	throws := findThrowStatements(f.Body)
	for _, t := range throws {
		if t.Value != nil && !err.Extends(t.Value.Type()) {
			p.report("Error type doesn't match expected type", t.Value.Loc())
		}
	}
}

// Type check all possible return points and see if they match
func typeCheckImplicitReturn(p *Parser, f *FunctionExpression) {
	f.returnType = f.Body.Type()

	if !typeCheckHappyReturn(p, f.Body, f.Body.Type()) {
		p.report("Mismatched types", f.Body.reportLoc())
	}

	tries := findTryExpressions(f.Body)
	for _, t := range tries {
		p.report(
			"Failable expressions are not allowed in functions without explicit returns",
			t.Loc(),
		)
	}
	throws := findThrowStatements(f.Body)
	for _, t := range throws {
		p.report("Cannot throw in functions without explicit returns", t.Loc())
	}
}

// Check all return points in a function body against an expected typing.
func typeCheckHappyReturn(p *Parser, body *Block, expected ExpressionType) bool {
	happy := getHappyType(expected)
	returns := findReturnStatements(body)
	ok := true
	if !expected.Extends(body.Type()) && !happy.Extends(body.Type()) {
		p.report("Type doesn't match expected type", body.reportLoc())
	}
	for _, r := range returns {
		returnType := getExitType(r)
		if !expected.Extends(returnType) && !happy.Extends(returnType) {
			ok = false
			p.report("Type doesn't match expected type", r.Value.Loc())
		}
	}
	return ok
}

func getExitType(e *Exit) ExpressionType {
	if e.Value == nil {
		return Primitive{NIL}
	}
	return e.Value.Type()
}

// Find all the return statements in a function body.
// Don't check inside nested functions.
func findReturnStatements(body *Block) []*Exit {
	results := []*Exit{}
	appendIfReturn := func(node Node) {
		if isReturnStatement(node) {
			results = append(results, node.(*Exit))
		}
	}
	body.Walk(appendIfReturn, isFunctionExpression)
	return results
}
func isReturnStatement(node Node) bool {
	exit, ok := node.(*Exit)
	if !ok {
		return false
	}
	return exit.Operator.Kind() == ReturnKeyword
}

// Find all the try expressions in a function body.
// Don't check inside nested functions.
func findTryExpressions(body *Block) []*TryExpression {
	results := []*TryExpression{}
	appendIfTry := func(node Node) {
		if n, ok := node.(*TryExpression); ok {
			results = append(results, n)
		}
	}
	body.Walk(appendIfTry, isFunctionExpression)
	return results
}

// Find all the throw statements in a function body.
// Don't check inside nested functions.
func findThrowStatements(body *Block) []*Exit {
	results := []*Exit{}
	appendIfTry := func(node Node) {
		if n, ok := node.(*Exit); ok && n.Operator.Kind() == ThrowKeyword {
			results = append(results, n)
		}
	}
	body.Walk(appendIfTry, isFunctionExpression)
	return results
}

func isFunctionExpression(node Node) bool {
	_, ok := node.(*FunctionExpression)
	return ok
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
