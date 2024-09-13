package checker

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type ObjectExpressionMember struct {
	Name  Identifier
	Value Expression
}

type ObjectExpression struct {
	Expr    Expression
	Members []ObjectExpressionMember
	typing  TypeAlias
	loc     tokenizer.Loc
}

func (o ObjectExpression) Loc() tokenizer.Loc   { return o.loc }
func (o ObjectExpression) Type() ExpressionType { return o.typing }

type ListExpression struct {
	Expr     Expression
	Elements []Expression
	typing   List
	loc      tokenizer.Loc
}

func (l ListExpression) Loc() tokenizer.Loc   { return l.loc }
func (l ListExpression) Type() ExpressionType { return l.typing }

func (c *Checker) checkObjectExpressionMember(node parser.Node) (ObjectExpressionMember, bool) {
	member, ok := node.(parser.TypedExpression)
	if !ok {
		return ObjectExpressionMember{}, false
	}
	if !member.Colon {
		c.report("Expected ':' between identifier and value", member.Loc())
	}
	token, ok := member.Expr.(parser.TokenExpression)
	if !ok {
		return ObjectExpressionMember{}, false
	}
	name, ok := c.checkToken(token, false).(Identifier)
	if !ok {
		return ObjectExpressionMember{}, false
	}
	value := c.checkExpression(member.Typing)
	return ObjectExpressionMember{name, value}, true
}

func (c *Checker) checkInstanciationExpression(node parser.InstanciationExpression) Expression {
	expr := c.checkExpression(node.Typing)
	typing, ok := expr.Type().(Type)
	if !ok {
		c.report("Type expected", node.Typing.Loc())
		return ObjectExpression{Expr: expr, loc: node.Loc()}
	}
	switch t := typing.Value.(type) {
	case TypeAlias:
		object, ok := t.Ref.(Object)
		if !ok {
			c.report("Object type expected", expr.Loc())
			return ObjectExpression{Expr: expr, typing: t, loc: node.Loc()}
		}
		c.pushScope(NewScope())
		defer c.dropScope()
		c.addTypeArgsToScope(nil, t.Params)
		members := checkObjectMembers(c, &object, node.Members)
		err := getMissingMembers(object.Members, members)
		if err != "" {
			c.report(fmt.Sprintf("Missing key(s) %v", err), node.Loc())
		}
		t.Ref = object
		return ObjectExpression{
			Expr:    expr,
			Members: members,
			typing:  t,
			loc:     node.Loc(),
		}
	case List:
		if len(node.Members) == 0 {
			return ListExpression{Expr: expr, Elements: nil, typing: t, loc: node.Loc()}
		}
		el := t.Element
		members := make([]Expression, len(node.Members))
		if alias, ok := t.Element.(TypeAlias); ok {
			m := c.checkExpression(node.Members[0])
			c.pushScope(NewScope())
			defer c.dropScope()
			c.addTypeArgsToScope(nil, alias.Params)
			el, _ = t.Element.build(c.scope, m.Type())
		}

		tail := node.Members[1:]
		for i, member := range tail {
			members[i] = c.checkExpression(member)
			if !el.Extends(members[i].Type()) {
				c.report("Type doesn't match", member.Loc())
			}
		}
		return ListExpression{Expr: expr, Elements: members, typing: t}
	default:
		c.report("Unexpected typing (expected object, list or sum type constructor)", expr.Loc())
		return ObjectExpression{Expr: expr, loc: node.Loc()}
	}
}

func checkObjectMembers(c *Checker, object *Object, nodes []parser.Node) []ObjectExpressionMember {
	var members []ObjectExpressionMember
	for _, node := range nodes {
		member, ok := c.checkObjectExpressionMember(node)
		if !ok {
			c.report("Expected member expression", node.Loc())
			continue
		}
		name := member.Name.Token.Text()
		expectedType, ok := object.Members[name].(Type).Value.build(c.scope, member.Value.Type())
		if !ok {
			c.report("Could not determine exact type", member.Value.Loc())
		}
		if !expectedType.Extends(member.Value.Type()) {
			c.report("Types don't match", node.Loc())
		}
		object.Members[name] = Type{expectedType}
		members = append(members, member)
	}
	return members
}

func getMissingMembers(expected map[string]ExpressionType, received []ObjectExpressionMember) string {
	membersSet := map[string]bool{}
	for name := range expected {
		membersSet[name] = true
	}
	for _, member := range received {
		delete(membersSet, member.Name.Text())
	}

	if len(membersSet) == 0 {
		return ""
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
	return msg
}
