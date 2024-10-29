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
	c.Callee.typeCheck(p)
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
	if alias.Name == "Map" {
		typeCheckMapInstanciation(p, c, alias)
		return
	}
	object, ok := alias.Ref.(Object)
	if !ok {
		p.report("Object type expected", c.Callee.Loc())
		c.typing = Primitive{UNKNOWN}
		return
	}
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	typeCheckTypeArgs(p, nil, alias.Params)

	args := c.Args.Expr.(*TupleExpression).Elements
	formatStructEntries(p, args)
	for _, arg := range args {
		entry := arg.(*Entry)
		var name string
		if entry.Key != nil {
			name = entry.Key.(*Identifier).Text()
		}
		expected := object.Members[name]
		if !expected.Extends(entry.Value.Type()) {
			p.report("Type doesn't match the expected one", arg.Loc())
		}
	}

	reportExcessMembers(p, object.Members, args)
	reportMissingMembers(p, object.Members, c.Args)

	c.typing = alias
}
func formatStructEntries(p *Parser, received []Expression) {
	for i := range received {
		received[i] = getFormattedStructEntry(p, received[i])
	}
}
func getFormattedStructEntry(p *Parser, received Expression) *Entry {
	entry, ok := received.(*Entry)
	if !ok {
		if param, ok := received.(*Param); ok {
			p.report("':' expected between key and value", param.Loc())
			return &Entry{Key: param.Identifier, Value: param.Complement}
		} else {
			p.report("Entry expected", received.Loc())
			return &Entry{Value: received}
		}
	}
	if _, ok := entry.Key.(*Identifier); !ok {
		p.report("Identifier expected", entry.Key.Loc())
		entry.Key = &Identifier{}
	}
	return entry
}
func reportExcessMembers(p *Parser, expected map[string]ExpressionType, received []Expression) {
	for _, arg := range received {
		namedArg, ok := arg.(*Entry)
		if !ok {
			continue
		}
		name := namedArg.Key.(*Identifier).Text()
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

func typeCheckMapInstanciation(p *Parser, c *CallExpression, t TypeAlias) {
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	typeCheckTypeArgs(p, nil, t.Params)
	args := c.Args.Expr.(*TupleExpression).Elements
	formatMapEntries(p, args)
	c.typing = getMapType(p, c)
	typeCheckMapEntries(p, args, c.typing.(TypeAlias).Ref.(Map))
}

func formatMapEntries(p *Parser, received []Expression) {
	for i := range received {
		received[i] = getFormattedMapEntry(p, received[i])
	}
}
func getFormattedMapEntry(p *Parser, received Expression) *Entry {
	entry, ok := received.(*Entry)
	if !ok {
		if param, ok := received.(*Param); ok {
			p.report("':' expected between key and value", param.Loc())
			return &Entry{Key: param.Identifier, Value: param.Complement}
		} else {
			p.report("Entry expected", received.Loc())
			return &Entry{Value: received}
		}
	}
	if _, ok := entry.Key.(*Identifier); ok {
		p.report("Literal or brackets expected", entry.Key.Loc())
		entry.Key = nil
	}
	return entry
}
func getMapType(p *Parser, c *CallExpression) ExpressionType {
	t := c.Callee.Type().(Type).Value.(TypeAlias)
	args := c.Args.Expr.(*TupleExpression).Elements
	var key, value ExpressionType
	var kk, vk bool
	if len(args) > 0 {
		key, kk = t.Params[0].build(p.scope, args[0].(*Entry).Key.Type())
	} else {
		key, kk = t.Params[0].build(p.scope, nil)
	}
	if len(args) > 0 {
		value, vk = t.Params[0].build(p.scope, args[0].(*Entry).Value.Type())
	} else {
		value, vk = t.Params[0].build(p.scope, nil)
	}
	if !kk || !vk {
		p.report(
			"Could not fully determine returned type, consider adding type arguments",
			c.Loc(),
		)
	}
	t.Params = append(t.Params[:0:0], t.Params...)
	t.Params[0].Value = key
	t.Params[1].Value = value
	m := t.Ref.(Map)
	m.Key = key
	m.Value = value
	t.Ref = m
	return t
}
func typeCheckMapEntries(p *Parser, entries []Expression, t Map) {
	for i := range entries {
		entry := entries[i].(*Entry)
		if entry.Key != nil {
			entry.Key.typeCheck(p)
			var key ExpressionType
			if b, ok := entry.Key.(*BracketedExpression); ok {
				key = b.Expr.Type()
			} else {
				key = entry.Key.Type()
			}
			if !t.Key.Extends(key) {
				p.report("Type doesn't match expected key type", entry.Key.Loc())
			}
		}
		if entry.Value != nil {
			entry.Value.typeCheck(p)
			if !t.Value.Extends(entry.Value.Type()) {
				p.report("Type doesn't match expected value type", entry.Value.Loc())
			}
		}
	}
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
		typeCheckTypeArgs(p, nil, alias.Params)
		el, _ = el.build(p.scope, first.Type())
	}

	for i := range elements[1:] {
		elements[i+1].typeCheck(p)
		if !el.Extends(elements[i+1].Type()) {
			p.report("Type doesn't match", elements[i+1].Loc())
		}
	}
}
