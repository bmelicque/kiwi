package parser

import "fmt"

// Callee(...Args)
type CallExpression struct {
	Callee Expression
	Args   *Params
	typing ExpressionType
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
	args := p.parseArguments()
	return &CallExpression{callee, args, nil}
}

func (c *CallExpression) typeCheck(p *Parser) {
	switch typing := c.Callee.Type().(type) {
	case Function:
		typeCheckFunctionCall(p, c)
	case Type:
		switch typing.Value.(type) {
		case TypeAlias:
			typeCheckStructInstanciation(p, c)
		case List:
			typeCheckListInstanciation(p, c)
		default:
			p.report("This type is not callable", c.Callee.Loc())
		}
	default:
		p.report(
			"This expression is not callable, type or function expected",
			c.Callee.Loc(),
		)
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

	params := function.Params.elements
	typeCheckFunctionArguments(p, *c.Args, params)
	validateArgumentsNumber(p, *c.Args, params)
	t, ok := function.Returned.build(p.scope, nil)
	if !ok {
		p.report(
			"Could not determine returned type (missing some type arguments)",
			c.Loc(),
		)
		c.typing = Primitive{UNKNOWN}
		return
	}
	c.typing = t
}

// Make sure that every parsed argument is compliant with the function's type
func typeCheckFunctionArguments(p *Parser, args Params, params []ExpressionType) {
	l := len(params)
	if len(args.Params) < len(params) {
		l = len(args.Params)
	}
	var ok bool
	for i := range args.Params[:l] {
		args.Params[i].Complement.typeCheck(p)
		received := args.Params[i].Complement.Type()
		params[i], ok = params[i].build(p.scope, received)
		if !ok {
			p.report("Could not determine exact type", args.Params[i].Loc())
		}
		if !params[i].Extends(received) {
			p.report("Types don't match", args.Params[i].Loc())
		}
	}
}

// Make sure that the correct number of arguments were passed to the function
func validateArgumentsNumber(p *Parser, args Params, params []ExpressionType) {
	if len(params) < len(args.Params) {
		loc := args.Params[len(params)].Loc()
		loc.End = args.Params[len(args.Params)-1].Loc().End
		p.report("Too many arguments", loc)
	}
	if len(params) > len(args.Params) {
		p.report("Missing argument(s)", args.Loc())
	}
}

// Parse a struct instanciation, like 'Object(key: value)'
func typeCheckStructInstanciation(p *Parser, c *CallExpression) {
	// next line ensured by calling function
	alias := c.Callee.Type().(Type).Value.(TypeAlias)
	object, ok := alias.Ref.(Object)
	if !ok {
		p.report("Object type expected", c.Callee.Loc())
		c.typing = Primitive{UNKNOWN}
		return
	}
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	p.addTypeArgsToScope(nil, alias.Params)
	c.Args.typeCheck(p)

	for _, arg := range c.Args.Params {
		if arg.kind == Argument {
			p.report("Named argument expected", arg.Loc())
			continue
		}
		var name string
		if arg.Identifier != nil {
			name = arg.Identifier.Text()
		}
		expected := object.Members[name]
		if !expected.Extends(arg.Complement.Type()) {
			p.report("Type doesn't match the expected one", arg.Loc())
		}
	}

	reportExcessMembers(p, object.Members, c.Args.Params)
	reportMissingMembers(p, object.Members, *c.Args)

	c.typing = alias
}
func reportExcessMembers(p *Parser, expected map[string]ExpressionType, received []Param) {
	for _, param := range received {
		name := param.Identifier.Text()
		if _, ok := expected[name]; ok {
			continue
		}
		p.report(
			fmt.Sprintf("Property '%v' doesn't exist on this type", name),
			param.Loc(),
		)
	}
}
func reportMissingMembers(p *Parser, expected map[string]ExpressionType, received Params) {
	membersSet := map[string]bool{}
	for name := range expected {
		membersSet[name] = true
	}
	for _, member := range received.Params {
		delete(membersSet, member.Identifier.Text())
	}

	if len(membersSet) == 0 {
		return
	}
	var msg string
	var i int
	for member := range membersSet {
		msg += fmt.Sprintf("'%v'", member)
		if i != len(membersSet)-1 {
			msg += ", "
		}
		i++
	}
	p.report(fmt.Sprintf("Missing key(s) %v", msg), received.loc)
}

func typeCheckListInstanciation(p *Parser, c *CallExpression) {
	// next line ensured by calling function
	c.typing = c.Callee.Type().(Type).Value.(List)
	if len(c.Args.Params) == 0 {
		return
	}
	first := c.Args.Params[0].Complement
	first.typeCheck(p)
	el := c.typing.(List).Element
	if alias, ok := el.(TypeAlias); ok {
		p.pushScope(NewScope(ProgramScope))
		defer p.dropScope()
		p.addTypeArgsToScope(nil, alias.Params)
		el, _ = el.build(p.scope, first.Type())
	}

	for i := range c.Args.Params[1:] {
		c.Args.Params[i+1].Complement.typeCheck(p)
		if !el.Extends(c.Args.Params[i+1].Complement.Type()) {
			p.report("Type doesn't match", c.Args.Params[i+1].Loc())
		}
	}
}
