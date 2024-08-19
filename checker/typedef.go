package checker

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type ObjectMemberDefinition struct {
	Name   Identifier
	Typing Expression
}

type ObjectDefinition struct {
	Members []ObjectMemberDefinition
	loc     tokenizer.Loc
}

func (expr ObjectDefinition) Loc() tokenizer.Loc { return expr.loc }

func (expr ObjectDefinition) Type() ExpressionType {
	value := Object{map[string]ExpressionType{}}
	for _, member := range expr.Members {
		value.Members[member.Name.Token.Text()] = member.Typing.Type()
	}
	return Type{value}
}

func (c *Checker) checkObjectDefinition(node parser.ObjectDefinition) ObjectDefinition {
	members := make([]ObjectMemberDefinition, len(node.Members))
	locs := map[string][]tokenizer.Loc{}

	for i, member := range node.Members {
		if member, ok := member.(parser.TypedExpression); ok {
			param := c.checkParam(member)
			members[i] = ObjectMemberDefinition{param.Identifier, param.Typing}
			locs[param.Identifier.Text()] = append(locs[param.Identifier.Text()], member.Expr.Loc())
		}
	}

	for name, locs := range locs {
		if len(locs) > 1 {
			for _, loc := range locs {
				c.report(fmt.Sprintf("Duplicate identifier '%v'", name), loc)
			}
		}
	}

	return ObjectDefinition{
		Members: members,
		loc:     node.Loc(),
	}
}
