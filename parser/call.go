package parser

import "fmt"

// Callee(...Args)
type CallExpression struct {
	Callee Expression
	Args   *Params
	typing ExpressionType
}

func (c CallExpression) Loc() Loc {
	return Loc{
		Start: c.Callee.Loc().Start,
		End:   c.Args.loc.End,
	}
}
func (c CallExpression) Type() ExpressionType { return c.typing }

// Primitive(value) | Object(key: value) | List(..values)
func parseInstanciation(p *Parser, expr Expression) *CallExpression {
	typing := expr.Type().(Type)

	switch getFromType(typing).(type) {
	case Primitive:
		return parsePrimitiveInstanciation(p, expr, typing)
	case TypeAlias:
		return parseStructInstanciation(p, expr, typing)
	case List:
		return parseListInstanciation(p, expr, typing)
	default:
		p.report("Unexpected typing", expr.Loc())
		return &CallExpression{Callee: expr}
	}
}

// Parse a primitive instanciation, like 'number(value)'.
// This is mostly used for sum types
func parsePrimitiveInstanciation(p *Parser, expr Expression, t Type) *CallExpression {
	args := p.getValidatedArguments(p.parseParenthesizedExpression())
	if len(args.Params) != 1 {
		p.report("Exactly 1 value expected", args.Loc())
	}
	var value Expression
	if len(args.Params) > 0 {
		value = args.Params[0].Complement
		if !getFromType(t).Extends(value.Type()) {
			p.report("Type doesn't match", value.Loc())
		}
	}
	return &CallExpression{
		Callee: expr,
		Args:   args,
		typing: getFinalType(t),
	}
}

// Parse a struct instanciation, like 'Object(key: value)'
func parseStructInstanciation(p *Parser, expr Expression, t Type) *CallExpression {
	alias := t.Value.(TypeAlias)
	object, ok := alias.Ref.(Object)
	if !ok {
		p.report("Object type expected", expr.Loc())
		return &CallExpression{Callee: expr, typing: t.Value}
	}
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	p.addTypeArgsToScope(nil, alias.Params)
	members := p.getValidatedNamedArguments(p.parseParenthesizedExpression())
	reportExcessMembers(p, object.Members, members.Params)
	reportMissingMembers(p, object.Members, *members)
	return &CallExpression{
		Callee: expr,
		Args:   members,
		typing: getFinalType(t),
	}
}
func reportExcessMembers(p *Parser, expected map[string]ExpressionType, received []Param) {
	for _, param := range received {
		name := param.Identifier.Text()
		_, ok := expected[name]
		if !ok {
			p.report(
				fmt.Sprintf("Property '%v' doesn't exist on this type", name),
				param.Loc(),
			)
		}
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

func parseListInstanciation(p *Parser, expr Expression, t Type) *CallExpression {
	args := p.getValidatedArguments(p.parseParenthesizedExpression())
	if len(args.Params) == 0 {
		return &CallExpression{Callee: expr, Args: args, typing: getFinalType(t)}
	}
	first := args.Params[0].Complement
	list := t.Value.(List)
	el := list.Element
	if alias, ok := list.Element.(TypeAlias); ok {
		p.pushScope(NewScope(ProgramScope))
		defer p.dropScope()
		p.addTypeArgsToScope(nil, alias.Params)
		el, _ = list.Element.build(p.scope, first.Type())
	}

	for _, arg := range args.Params[1:] {
		if !el.Extends(arg.Complement.Type()) {
			p.report("Type doesn't match", arg.Loc())
		}
	}
	return &CallExpression{
		Callee: expr,
		Args:   args,
		typing: getFinalType(t),
	}
}

func getFromType(t Type) ExpressionType {
	if constructor, ok := t.Value.(Constructor); ok {
		return constructor.From
	}
	return t.Value
}
func getFinalType(t Type) ExpressionType {
	if constructor, ok := t.Value.(Constructor); ok {
		return constructor.To
	}
	return t.Value
}

// Parse a function call. The expected form is `callee(..arguments)`
func parseFunctionCall(p *Parser, callee Expression) *CallExpression {
	args := p.getValidatedArguments(p.parseParenthesizedExpression())
	function, ok := callee.Type().(Function)
	if !ok {
		return &CallExpression{callee, args, nil}
	}

	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	for _, param := range function.TypeParams {
		// TODO: get declared location
		p.scope.Add(param.Name, Loc{}, Type{param})
	}

	params := function.Params.elements
	validateFunctionArguments(p, *args, params)
	t, ok := function.Returned.build(p.scope, nil)
	if !ok {
		p.report("Could not determine exact type", callee.Loc())
	}
	validateArgumentsNumber(p, *args, params)
	return &CallExpression{callee, args, t}
}

// Make sure that every parsed argument is compliant with the function's type
func validateFunctionArguments(p *Parser, args Params, params []ExpressionType) {
	l := len(params)
	if len(args.Params) < len(params) {
		l = len(args.Params)
	}
	var ok bool
	for i, element := range args.Params[:l] {
		received := element.Complement.Type()
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
