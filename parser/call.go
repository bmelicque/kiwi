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
	if args.Expr != nil {
		args.Expr = makeTuple(args.Expr)
	}
	return &CallExpression{callee, args, nil}
}

func (c *CallExpression) typeCheck(p *Parser) {
	c.Callee.typeCheck(p)
	switch c.Callee.Type().(type) {
	case Function:
		typeCheckFunctionCall(p, c)
	default:
		p.report("Function expected", c.Callee.Loc())
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
		p.report(
			"Could not determine returned type (missing some type arguments)",
			c.Loc(),
		)
		c.typing = Unknown{}
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
	var ok bool
	for i, element := range args.Elements[:l] {
		element.typeCheck(p)
		if _, ok := element.(*Param); ok {
			p.report("Single expression expected", element.Loc())
			continue
		}
		received := element.Type()
		params[i], ok = params[i].build(p.scope, received)
		if !ok {
			p.report("Could not determine exact type", element.Loc())
		}
		if !params[i].Extends(received) {
			p.report("Types don't match", element.Loc())
		}
	}
}

// Make sure that the correct number of arguments were passed to the function
func validateArgumentsNumber(p *Parser, args *TupleExpression, params []ExpressionType) {
	if len(params) < len(args.Elements) {
		loc := args.Elements[len(params)].Loc()
		loc.End = args.Elements[len(args.Elements)-1].Loc().End
		p.report("Too many arguments", loc)
	}
	if len(params) > len(args.Elements) {
		p.report("Missing argument(s)", args.Loc())
	}
}
