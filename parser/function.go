package parser

type FunctionExpression struct {
	TypeParams *BracketedExpression     // contains *TupleExpression of *Param
	Params     *ParenthesizedExpression // contains *TupleExpression
	Explicit   Expression               // Explicit return type (if any)
	Body       *Block
	returnType ExpressionType
	canBeAsync bool
}

func (f *FunctionExpression) getChildren() []Node {
	children := []Node{f.Params}
	if f.Explicit != nil {
		children = append(children, f.Explicit)
	}
	if f.Body != nil {
		children = append(children, f.Body)
	}
	return children
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
		f.TypeParams.getGenerics()
	}
	tuple, _ := f.Params.Type().(Tuple)
	return Function{tp, &tuple, f.returnType, f.canBeAsync}
}

func (f *FunctionExpression) typeCheck(p *Parser) {
	p.pushScope(NewScope(FunctionScope))
	defer p.dropScope()

	typeCheckTypeParams(p, f.TypeParams)

	if f.Params != nil && f.Params.Expr != nil {
		addParamsToScope(p, f.Params.Expr.(*TupleExpression).Elements)
	}
	f.Body.typeCheck(p)

	if f.Explicit != nil {
		typeCheckExplicitReturn(p, f)
	} else {
		typeCheckImplicitReturn(p, f)
	}

	f.canBeAsync = containsAsync(f)
}

type FunctionTypeExpression struct {
	TypeParams *BracketedExpression
	Params     *ParenthesizedExpression // Contains *TupleExpression
	Expr       Expression
}

func (f *FunctionTypeExpression) getChildren() []Node {
	children := []Node{f.Params}
	if f.Expr != nil {
		children = append(children, f.Expr)
	}
	return children
}

func (f *FunctionTypeExpression) Loc() Loc {
	var start, end Position
	if f.TypeParams != nil {
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
		tp = f.TypeParams.getGenerics()
	}
	elements := f.Params.Expr.(*TupleExpression).Elements
	p := Tuple{make([]ExpressionType, len(elements))}
	for i, param := range elements {
		t, _ := param.Type().(Type)
		p.elements[i] = t.Value
	}
	var ret ExpressionType = Unknown{}
	if f.Expr != nil {
		t, ok := f.Expr.Type().(Type)
		if ok {
			ret = t.Value
		}
	}
	return Type{Function{TypeParams: tp, Params: &p, Returned: ret}}
}

func (f *FunctionTypeExpression) typeCheck(p *Parser) {
	if f.TypeParams != nil {
		typeCheckTypeParams(p, f.TypeParams)
	}
	tuple := f.Params.Expr.(*TupleExpression)
	for i := range tuple.Elements {
		tuple.Elements[i].typeCheck(p)
		if _, ok := tuple.Elements[i].Type().(Type); !ok {
			p.report("Type expected", tuple.Elements[i].Loc())
		}
	}

	if f.Expr == nil {
		return
	}
	if _, ok := f.Expr.Type().(Type); !ok {
		p.report("Type expected", f.Expr.Loc())
	}
}

func (p *Parser) parseFunctionExpression(typeParams *BracketedExpression) Expression {
	p.pushScope(NewScope(FunctionScope))
	defer p.dropScope()

	if typeParams != nil {
		validateTypeParams(p, typeParams)
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
	if _, ok := expr.(*Param); !ok {
		p.report("Parameter expected: (name Type)", expr.Loc())
		return
	}
}

// Type check all possible return points against the explicit return type.
// Also check possible failure points.
func typeCheckExplicitReturn(p *Parser, f *FunctionExpression) {
	explicit := f.Explicit.Type()
	if t, ok := explicit.(Type); ok {
		explicit = t.Value
	} else {
		explicit = Unknown{}
	}
	f.returnType = explicit

	typeCheckHappyReturn(p, f.Body, explicit)

	err := getErrorType(explicit)
	tries := findTryExpressions(f.Body)
	for _, t := range tries {
		if !err.Extends(getErrorType(t.Operand.Type())) {
			p.report("Error type doesn't match expected type", t.Operand.Loc())
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

	returns := findReturnStatements(f.Body)
	for _, r := range returns {
		p.report("Cannot return in functions without explicit returns", r.Loc())
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
		return Nil{}
	}
	return e.Value.Type()
}

// Find all the return statements in a function body.
// Don't check inside nested functions.
func findReturnStatements(body *Block) []*Exit {
	results := []*Exit{}
	Walk(body, func(n Node, skip func()) {
		if isFunctionExpression(n) {
			skip()
		}
		if isReturnStatement(n) {
			results = append(results, n.(*Exit))
		}
	})
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
func findTryExpressions(body *Block) []*UnaryExpression {
	results := []*UnaryExpression{}
	Walk(body, func(n Node, skip func()) {
		if isFunctionExpression(n) {
			skip()
		}
		if n, ok := n.(*UnaryExpression); ok && n.Operator.Kind() == TryKeyword {
			results = append(results, n)
		}
	})
	return results
}

// Find all the throw statements in a function body.
// Don't check inside nested functions.
func findThrowStatements(body *Block) []*Exit {
	results := []*Exit{}
	Walk(body, func(n Node, skip func()) {
		if isFunctionExpression(n) {
			skip()
		}
		if n, ok := n.(*Exit); ok && n.Operator.Kind() == ThrowKeyword {
			results = append(results, n)
		}
	})
	return results
}

func isFunctionExpression(node Node) bool {
	_, ok := node.(*FunctionExpression)
	return ok
}

func addParamsToScope(p *Parser, tuple []Expression) {
	for _, expr := range tuple {
		addParamToScope(p, expr)
	}
}

func addParamToScope(p *Parser, expr Expression) {
	param, ok := expr.(*Param)
	if !ok {
		return
	}
	if param.Complement == nil {
		p.report("Typing expected", param.Loc())
		p.scope.Add(param.Identifier.Text(), param.Loc(), Unknown{})
		return
	}
	if _, ok := param.Complement.Type().(Type); !ok {
		p.report("Typing expected", param.Loc())
		p.scope.Add(param.Identifier.Text(), param.Loc(), Unknown{})
		return
	}
	typing, _ := param.Complement.Type().(Type)
	p.scope.Add(param.Identifier.Text(), param.Loc(), typing.Value)
}

func containsAsync(f *FunctionExpression) bool {
	var async bool
	Walk(f.Body, func(n Node, skip func()) {
		if async || isFunctionExpression(n) {
			skip()
		}
		expr, ok := n.(Expression)
		if !ok {
			return
		}
		f, ok := expr.Type().(Function)
		if !ok {
			return
		}
		if f.Async {
			async = true
			skip()
		}
	})
	return async
}
