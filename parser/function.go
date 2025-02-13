package parser

type FunctionExpression struct {
	TypeParams *BracketedExpression     // contains *TupleExpression of *Param
	Params     *ParenthesizedExpression // contains *TupleExpression
	Arrow      Token
	Explicit   Expression // Explicit return type (if any)
	Body       *Block
	typing     Function
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
	if f.TypeParams != nil {
		loc.Start = f.TypeParams.loc.Start
	}
	if f.Body != nil {
		loc.End = f.Body.Loc().End
	} else if f.Explicit != nil {
		loc.End = f.Explicit.Loc().End
	} else {
		loc.End = f.Arrow.Loc().End
	}
	return loc
}

func (f *FunctionExpression) Type() ExpressionType { return f.typing }

func (f *FunctionExpression) typeCheck(p *Parser) {
	typeCheckFunctionExpression(p, f, func(params *TupleExpression) {
		for _, param := range params.Elements {
			if _, ok := param.(*Param); !ok {
				p.error(param, ParameterExpected)
			}
		}
		addParamsToScope(p, params.Elements)
	})
}

func typeCheckFunctionExpression(p *Parser, f *FunctionExpression, paramHandler func(params *TupleExpression)) {
	p.pushScope(NewScope(FunctionScope))
	defer p.dropScope()

	typeCheckTypeParams(p, f.TypeParams)

	if f.Params != nil && f.Params.Expr != nil {
		paramHandler(f.Params.Expr.(*TupleExpression))
	}
	f.Body.typeCheck(p)

	if f.Explicit != nil {
		f.Explicit.typeCheck(p)
		f.typeCheckBodyExplicit(p)
	} else {
		typeCheckImplicitReturn(p, f)
	}

	f.canBeAsync = containsAsync(f)
	f.typing = getFunctionType(f)
}

func typeCheckHOF(p *Parser, f *FunctionExpression, expected Tuple) {
	typeCheckFunctionExpression(p, f, func(params *TupleExpression) {
		l := checkHOFParamsLength(p, expected, params)
		for i := 0; i < l; i++ {
			expectedType := expected.Elements[i]
			typeCheckHOFParam(p, params.Elements[i], expectedType)
			addHOFParamToScope(p, params.Elements[i], expectedType)
		}
	})
}

// returns the number of elements that can be safely iterated
func checkHOFParamsLength(p *Parser, expected Tuple, received *TupleExpression) int {
	le := len(expected.Elements)
	lr := len(received.Elements)
	if le < lr {
		p.error(received, TooManyElements, le, lr)
		return le
	} else if le > lr {
		p.error(received, MissingElements, le, lr)
		return lr
	}
	return le
}

func typeCheckHOFParam(p *Parser, expr Expression, expectedType ExpressionType) {
	var received ExpressionType
	param, ok := expr.(*Param)
	if !ok {
		return
	}
	received = param.Complement.Type()
	t, ok := received.(Type)
	if !ok {
		return
	}
	if !expectedType.Extends(t.Value) {
		p.error(expr, CannotAssignType, expectedType, t.Value)
	}
}

func addHOFParamToScope(p *Parser, param Expression, expected ExpressionType) {
	switch param := param.(type) {
	case *Param:
		addParamToScope(p, param)
	case *Identifier:
		p.scope.Add(param.Text(), param.Loc(), expected)
	default:
		panic("param or identifier expected")
	}
}

func getFunctionType(f *FunctionExpression) Function {
	returned := getFunctionReturnedType(f)
	typeParams := []Generic{}
	if f.TypeParams != nil {
		typeParams = f.TypeParams.getGenerics()
	}
	params := getFunctionParamsType(f)
	return Function{typeParams, &params, returned, f.canBeAsync}
}

func getFunctionParamsType(f *FunctionExpression) Tuple {
	if len(f.Params.Expr.(*TupleExpression).Elements) == 0 {
		return Tuple{[]ExpressionType{}}
	}
	t := f.Params.Type()
	if tu, ok := t.(Tuple); ok {
		return tu
	} else {
		return Tuple{[]ExpressionType{t}}
	}
}

func getFunctionReturnedType(f *FunctionExpression) ExpressionType {
	if f.Explicit == nil {
		return f.Body.Type()
	}
	if i, ok := Unwrap(f.Explicit).(*Identifier); ok && i.Text() == "_" {
		return Void{}
	}
	t, ok := f.Explicit.Type().(Type)
	if !ok {
		return Invalid{}
	}
	return t.Value
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
		p.Elements[i] = t.Value
	}
	var ret ExpressionType = Invalid{}
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
		t, ok := tuple.Elements[i].Type().(Type)
		if !ok {
			p.error(tuple.Elements[i], TypeExpected)
		} else if t.Value == (Void{}) {
			p.error(tuple.Elements[i], VoidAssignment)
		}
	}

	if f.Expr == nil {
		return
	}
	if _, ok := f.Expr.Type().(Type); !ok {
		p.error(f.Expr, TypeExpected)
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
			paren.Expr = MakeTuple(paren.Expr)
		}
		old := p.allowBraceParsing
		p.allowBraceParsing = false
		expr := p.parseBinaryExpression()
		p.allowBraceParsing = old
		return &FunctionTypeExpression{
			TypeParams: typeParams,
			Params:     paren,
			Expr:       expr,
		}
	case FatArrow:
		arrow := p.Consume() // =>
		if paren != nil {
			paren.Expr = MakeTuple(paren.Expr)
			validateFunctionParams(p, paren)
		}
		outerBrace := p.allowBraceParsing
		outerEmpty := p.allowEmptyExpr
		p.allowBraceParsing = false
		p.allowEmptyExpr = true
		explicit := p.parseBinaryExpression()
		p.allowBraceParsing = outerBrace
		p.allowEmptyExpr = outerEmpty
		body := p.parseBlock()
		return &FunctionExpression{
			TypeParams: typeParams,
			Params:     paren,
			Arrow:      arrow,
			Explicit:   explicit,
			Body:       body,
		}
	default:
		return paren
	}
}

// Validate all of a function params' structures
func validateFunctionParams(p *Parser, node *ParenthesizedExpression) {
	tuple := node.Expr.(*TupleExpression)
	for _, element := range tuple.Elements {
		validateFunctionParam(p, element)
	}

	tuple.reportDuplicatedParams(p)
}

// Validate a function param's structure
func validateFunctionParam(p *Parser, expr Expression) {
	switch expr := expr.(type) {
	case *Param, *Identifier:
	default:
		p.error(expr, ParameterExpected)
	}
}

// Type check all possible return points against the explicit return type.
// Also check possible failure points.
func (f *FunctionExpression) typeCheckBodyExplicit(p *Parser) {
	i, ok := Unwrap(f.Explicit).(*Identifier)
	isVoid := ok && i.Text() == "_"
	if _, ok := f.Explicit.Type().(Type); !ok && !isVoid {
		p.error(f.Explicit, TypeExpected)
		return
	}
	var explicit ExpressionType
	if isVoid {
		explicit = Type{Void{}}
	} else {
		explicit = f.Explicit.Type()
	}
	typeCheckReturnsExplicit(p, explicit, f.Body)
	err := f.getExplicitErrorType()
	checkFunctionTries(p, err, findTryExpressions(f.Body))
	checkFunctionThrows(p, err, findThrowStatements(f.Body))
}
func (f *FunctionExpression) getExplicitErrorType() ExpressionType {
	t, ok := f.Explicit.Type().(Type)
	if !ok {
		return nil
	}
	return getErrorType(t.Value)
}
func checkFunctionTries(p *Parser, err ExpressionType, tries []*UnaryExpression) {
	for _, t := range tries {
		if !err.Extends(getErrorType(t.Operand.Type())) {
			p.error(t.Operand, CannotAssignType, err, t.Operand.Type())
		}
	}
}
func checkFunctionThrows(p *Parser, err ExpressionType, throws []*Exit) {
	for _, t := range throws {
		if t.Value != nil && !err.Extends(t.Value.Type()) {
			p.error(t.Value, CannotAssignType, err, t.Value)
		}
	}
}

// Type check all possible return points and see if they match
func typeCheckImplicitReturn(p *Parser, f *FunctionExpression) {
	returns := findReturnStatements(f.Body)
	for _, r := range returns {
		p.error(r, IllegalReturn)
	}

	tries := findTryExpressions(f.Body)
	for _, t := range tries {
		p.error(t, IllegalResult)
	}
	throws := findThrowStatements(f.Body)
	for _, t := range throws {
		p.error(t, IllegalThrow)
	}
}

// Check all return points in a function body against an expected typing.
func typeCheckReturnsExplicit(p *Parser, explicit ExpressionType, body *Block) {
	t, ok := explicit.(Type)
	if !ok {
		return
	}
	expected := getHappyType(t.Value)
	returns := findReturnStatements(body)
	bodyType := body.Type()
	if !IsExiting(body.Last()) && !expected.Extends(bodyType) {
		p.error(body.reportedNode(), CannotAssignType, expected, bodyType)
	}
	for _, r := range returns {
		returnType := getExitType(r)
		if !expected.Extends(returnType) {
			p.error(r.Value, CannotAssignType, expected, returnType)
		}
	}
}

func getExitType(e *Exit) ExpressionType {
	if e.Value == nil {
		return Void{}
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
		p.error(param, TypeExpected)
		p.scope.Add(param.Identifier.Text(), param.Loc(), Invalid{})
		return
	}
	param.Complement.typeCheck(p)
	if _, ok := param.Complement.Type().(Type); !ok {
		p.error(param, TypeExpected)
		p.scope.Add(param.Identifier.Text(), param.Loc(), Invalid{})
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
