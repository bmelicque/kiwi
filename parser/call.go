package parser

import "fmt"

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
	typeCheckFunctionArguments(p, c.Args.Expr.(*TupleExpression), params)
	validateArgumentsNumber(p, c.Args.Expr.(*TupleExpression), params)
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

	args := c.Args.Expr.(*TupleExpression).Elements
	for _, arg := range args {
		namedArg, ok := arg.(*Param)
		if !ok {
			p.report("Named argument expected", arg.Loc())
			continue
		}
		var name string
		if namedArg.Identifier != nil {
			name = namedArg.Identifier.Text()
		}
		expected := object.Members[name]
		if !expected.Extends(namedArg.Complement.Type()) {
			p.report("Type doesn't match the expected one", arg.Loc())
		}
	}

	reportExcessMembers(p, object.Members, args)
	reportMissingMembers(p, object.Members, c.Args)

	c.typing = alias
}
func reportExcessMembers(p *Parser, expected map[string]ExpressionType, received []Expression) {
	for _, arg := range received {
		namedArg, ok := arg.(*Param)
		if !ok {
			continue
		}
		name := namedArg.Identifier.Text()
		if _, ok := expected[name]; ok {
			continue
		}
		p.report(
			fmt.Sprintf("Property '%v' doesn't exist on this type", name),
			arg.Loc(),
		)
	}
}
func reportMissingMembers(
	p *Parser,
	expected map[string]ExpressionType,
	received *ParenthesizedExpression,
) {
	membersSet := map[string]bool{}
	for name := range expected {
		membersSet[name] = true
	}
	for _, member := range received.Expr.(*TupleExpression).Elements {
		if named, ok := member.(*Param); ok {
			delete(membersSet, named.Identifier.Text())
		}
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
	elements := c.Args.Expr.(*TupleExpression).Elements
	if len(elements) == 0 {
		return
	}
	first := elements[0]
	first.typeCheck(p)
	el := c.typing.(List).Element
	if alias, ok := el.(TypeAlias); ok {
		p.pushScope(NewScope(ProgramScope))
		defer p.dropScope()
		p.addTypeArgsToScope(nil, alias.Params)
		el, _ = el.build(p.scope, first.Type())
	}

	for i := range elements[1:] {
		elements[i+1].typeCheck(p)
		if !el.Extends(elements[i+1].Type()) {
			p.report("Type doesn't match", elements[i+1].Loc())
		}
	}
}
