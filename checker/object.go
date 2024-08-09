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
	Typing  Expression
	Members []ObjectExpressionMember
	loc     tokenizer.Loc
}

func (o ObjectExpression) Loc() tokenizer.Loc   { return o.loc }
func (o ObjectExpression) Type() ExpressionType { return o.Typing.Type() }

func getObjectExpressionTyping(expr Expression) (Object, bool) {
	typing, ok := expr.Type().(Type)
	if !ok {
		return Object{}, false
	}
	ref, ok := typing.Value.(TypeRef)
	if !ok {
		return Object{}, false
	}
	object, ok := ref.Ref.(Object)
	if !ok {
		return Object{}, false
	}
	return object, true
}

func (c *Checker) checkObjectExpressionMember(node parser.Node) (ObjectExpressionMember, bool) {
	member, ok := node.(parser.TypedExpression)
	if !ok {
		return ObjectExpressionMember{}, false
	}
	token, ok := member.Expr.(parser.TokenExpression)
	if !ok {
		return ObjectExpressionMember{}, false
	}
	name, ok := c.checkToken(&token, false).(Identifier)
	if !ok {
		return ObjectExpressionMember{}, false
	}

	value := c.CheckExpression(member.Typing)
	return ObjectExpressionMember{name, value}, true
}

func (c *Checker) checkObjectExpression(expr parser.ObjectExpression) ObjectExpression {
	typing := c.CheckExpression(expr.Typing)
	object, ok := getObjectExpressionTyping(typing)
	if !ok {
		c.report("Expected object type", expr.Typing.Loc())
	}

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
		if ok && !object.Members[name].Match(member.Value.Type()) {
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
