package parser

// Callee(...Args)
type CallExpression struct {
	Callee Expression
	Args   *ParenthesizedExpression // contains a *TupleExpression
	typing ExpressionType
}

func (c *CallExpression) getChildren() []Node {
	children := []Node{c.Callee}
	if c.Args != nil {
		children = append(children, c.Args)
	}
	return children
}

func (c *CallExpression) Loc() Loc {
	return Loc{
		Start: c.Callee.Loc().Start,
		End:   c.Args.loc.End,
	}
}
func (c *CallExpression) Type() ExpressionType { return c.typing }

// Parse a call expression.
// It can be either a function call or an instanciation.
func parseCallExpression(p *Parser, callee Expression) *CallExpression {
	args := p.parseParenthesizedExpression()
	args.Expr = makeTuple(args.Expr)
	return &CallExpression{callee, args, nil}
}

func (c *CallExpression) typeCheck(p *Parser) {
	c.Callee.typeCheck(p)
	switch c.Callee.Type().(type) {
	case Function:
		typeCheckFunctionCall(p, c)
	default:
		p.error(c.Callee, FunctionExpressionExpected)
		c.Args.typeCheck(p)
	}
}

func typeCheckFunctionCall(p *Parser, c *CallExpression) {
	function := c.Callee.Type().(Function)

	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	for _, param := range function.TypeParams {
		// TODO: get declared location
		p.scope.Add(param.Name, Loc{}, Type{param})
	}

	params := function.Params.Elements
	typeCheckFunctionArguments(p, c.Args.Expr.(*TupleExpression), params)
	validateArgumentsNumber(p, c.Args.Expr.(*TupleExpression), params)
	t, ok := function.Returned.build(p.scope, nil)
	if !ok {
		p.error(c, MissingTypeArgs)
		c.typing = Invalid{}
		return
	}
	c.typing = t
}

// Make sure that every parsed argument is compliant with the function's type
func typeCheckFunctionArguments(p *Parser, args *TupleExpression, params []ExpressionType) {
	l := len(params)
	if len(args.Elements) < len(params) {
		l = len(args.Elements)
	}
	for i, element := range args.Elements[:l] {
		typeCheckFunctionArgument(p, &params[i], element)
	}
}

func typeCheckFunctionArgument(p *Parser, expected *ExpressionType, received Expression) {
	if f, ok := received.(*FunctionExpression); ok {
		typeCheckFunctionHOFArgument(p, *expected, f)
		return
	}
	received.typeCheck(p)
	if _, ok := received.(*Param); ok {
		p.error(received, ExpressionExpected)
		return
	}
	t := received.Type()
	built, ok := (*expected).build(p.scope, t)
	*expected = built
	if !ok {
		p.error(received, MissingTypeArgs)
	}
	if !built.Extends(t) {
		p.error(received, CannotAssignType, built, t)
	}
}

func typeCheckFunctionHOFArgument(p *Parser, expected ExpressionType, received *FunctionExpression) {
	if e, ok := expected.(Function); ok {
		typeCheckHOF(p, received, *e.Params)
	} else {
		received.typeCheck(p)
		p.error(received, CannotAssignType, expected, received.typing)
	}
}

// Make sure that the correct number of arguments were passed to the function
func validateArgumentsNumber(p *Parser, args *TupleExpression, params []ExpressionType) {
	if len(params) < len(args.Elements) {
		p.error(args, TooManyElements, len(params), len(args.Elements))
	}
	if len(params) > len(args.Elements) {
		p.error(args, MissingElements, len(params), len(args.Elements))
	}
}
