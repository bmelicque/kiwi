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
	Typing   Expression
	TypeArgs *TupleExpression
	Members  []ObjectExpressionMember
	loc      tokenizer.Loc
}

func (o ObjectExpression) Loc() tokenizer.Loc   { return o.loc }
func (o ObjectExpression) Type() ExpressionType { return o.Typing.Type().(Type).Value }

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

func (c *Checker) checkObjectExpression(expr parser.ObjectExpression) ObjectExpression {
	typing := c.checkExpression(expr.Typing)
	object, ok := handleGenericType(c, typing, expr.TypeArgs)

	var members []ObjectExpressionMember
	membersSet := map[string]bool{}
	for _, node := range expr.Members {
		member, k := c.checkObjectExpressionMember(node)
		if !k {
			c.report("Expected member expression", node.Loc())
			continue
		}
		name := member.Name.Token.Text()
		membersSet[name] = true
		expectedType := object.Members[name].(Type).Value.build(c.scope, member.Value.Type())
		if ok && !expectedType.Match(member.Value.Type()) {
			c.report("Types don't match", node.Loc())
		}
		members = append(members, member)
	}

	if ok {
		for name := range object.Members {
			if _, ok := membersSet[name]; !ok {
				c.report(fmt.Sprintf("Missing key '%v'", name), expr.Loc())
			}
		}
	}

	return ObjectExpression{
		Typing:  typing,
		Members: members,
	}
}

func handleGenericType(c *Checker, expr Expression, typeArgs *parser.BracketedExpression) (Object, bool) {
	t, ok := expr.Type().(Type)
	if !ok {
		c.report("Typing expected", expr.Loc())
		return Object{}, false
	}
	alias, ok := t.Value.(TypeAlias)
	if !ok {
		c.report("Type alias expected", expr.Loc())
		return Object{}, false
	}
	args := checkTypeArgs(c, typeArgs)
	c.addTypeArgsToScope(args, alias.Params)
	object, ok := alias.Ref.(Object)
	if !ok {
		c.report("Object reference expected", expr.Loc())
	}
	return object, ok
}
