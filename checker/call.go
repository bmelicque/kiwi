package checker

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

type CallExpression struct {
	Callee Expression
	Args   Params
	typing ExpressionType
}

func (c CallExpression) Loc() parser.Loc {
	loc := c.Args.loc
	if c.Callee != nil {
		loc.Start = c.Callee.Loc().Start
	}
	return loc
}

func (c CallExpression) Type() ExpressionType { return c.typing }

func (c *Checker) checkCallExpression(expr parser.CallExpression) Expression {
	callee := c.checkExpression(expr.Callee)
	_, isType := callee.Type().(Type)
	if isType {
		return checkInstanciation(c, expr)
	}

	return checkFunctionCall(c, callee, expr.Args)
}

// function(..args)
func checkFunctionCall(c *Checker, callee Expression, node parser.ParenthesizedExpression) Expression {
	function, ok := callee.Type().(Function)
	if !ok {
		args := c.checkArguments(node)
		return CallExpression{callee, args, nil}
	}

	c.pushScope(NewScope(ProgramScope))
	defer c.dropScope()
	for _, param := range function.TypeParams {
		// TODO: get declared location
		c.scope.Add(param.Name, parser.Loc{}, Type{param})
	}

	args := c.checkArguments(node)
	params := function.Params.elements
	checkFunctionArgs(c, args, params)
	t, ok := function.Returned.build(c.scope, nil)
	if !ok {
		c.report("Could not determine exact type", callee.Loc())
	}
	checkFunctionArgsNumber(c, args, params)
	return CallExpression{callee, args, t}
}
func checkFunctionArgsNumber(c *Checker, args Params, params []ExpressionType) {
	if len(params) < len(args.Params) {
		loc := args.Params[len(params)].Loc()
		loc.End = args.Params[len(args.Params)-1].Loc().End
		c.report("Too many arguments", loc)
	}
	if len(params) > len(args.Params) {
		c.report("Missing argument(s)", args.Loc())
	}
}
func checkFunctionArgs(c *Checker, args Params, params []ExpressionType) {
	l := len(params)
	if len(args.Params) < len(params) {
		l = len(args.Params)
	}
	var ok bool
	for i, element := range args.Params[:l] {
		received := element.Complement.Type()
		params[i], ok = params[i].build(c.scope, received)
		if !ok {
			c.report("Could not determine exact type", element.Loc())
		}
		if !params[i].Extends(received) {
			c.report("Types don't match", element.Loc())
		}
	}
}

// Primitive(value) | Object(key: value) | List(..values)
func checkInstanciation(c *Checker, node parser.CallExpression) Expression {
	expr := c.checkExpression(node.Callee)

	typing := expr.Type().(Type)
	from := typing
	if constructor, ok := typing.Value.(Constructor); ok {
		from.Value = constructor.From
	}

	switch t := from.Value.(type) {
	case Primitive:
		args := c.checkArguments(node.Args)
		if len(args.Params) != 1 {
			c.report("Exactly 1 value expected", node.Loc())
		}
		var value Expression
		if len(args.Params) > 0 {
			value = args.Params[0].Complement
			if !t.Extends(value.Type()) {
				c.report("Type doesn't match", value.Loc())
			}
		}
		return CallExpression{
			Callee: expr,
			Args:   args,
			typing: getFinalType(typing),
		}
	case TypeAlias:
		object, ok := t.Ref.(Object)
		if !ok {
			c.report("Object type expected", expr.Loc())
			return CallExpression{Callee: expr, typing: t}
		}
		c.pushScope(NewScope(ProgramScope))
		defer c.dropScope()
		c.addTypeArgsToScope(nil, t.Params)
		members := c.checkNamedArguments(node.Args)
		reportExcessMembers(c, object.Members, members.Params)
		reportMissingMembers(c, object.Members, members)
		t.Ref = object
		return CallExpression{
			Callee: expr,
			Args:   members,
			typing: getFinalType(typing),
		}
	case List:
		args := c.checkArguments(node.Args)
		if len(args.Params) == 0 {
			return CallExpression{Callee: expr, Args: args, typing: t}
		}
		first := args.Params[0].Complement
		el := t.Element
		if alias, ok := t.Element.(TypeAlias); ok {
			c.pushScope(NewScope(ProgramScope))
			defer c.dropScope()
			c.addTypeArgsToScope(nil, alias.Params)
			el, _ = t.Element.build(c.scope, first.Type())
		}

		for _, arg := range args.Params[1:] {
			if !el.Extends(arg.Complement.Type()) {
				c.report("Type doesn't match", arg.Loc())
			}
		}
		return CallExpression{
			Callee: expr,
			Args:   args,
			typing: getFinalType(typing),
		}
	default:
		c.report("Unexpected typing (expected object, list or sum type constructor)", expr.Loc())
		return CallExpression{Callee: expr}
	}
}

func reportExcessMembers(c *Checker, expected map[string]ExpressionType, received []Param) {
	for _, param := range received {
		name := param.Identifier.Text()
		_, ok := expected[name]
		if !ok {
			c.report(fmt.Sprintf("Property '%v' doesn't exist on this type", name), param.loc)
		}
	}
}
func reportMissingMembers(c *Checker, expected map[string]ExpressionType, received Params) {
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
	c.report(fmt.Sprintf("Missing key(s) %v", msg), received.loc)
}

func getFinalType(t Type) ExpressionType {
	if constructor, ok := t.Value.(Constructor); ok {
		return constructor.To
	}
	return t
}
