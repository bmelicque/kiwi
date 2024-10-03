package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type SumTypeMember struct {
	Name   Identifier
	Typing Expression
}

type SumType struct {
	Members []SumTypeMember
	typing  Type // Type{Sum}
	loc     tokenizer.Loc
}

func (s SumType) Loc() tokenizer.Loc   { return s.loc }
func (s SumType) Type() ExpressionType { return s.typing }

func (c *Checker) checkSumType(node parser.SumType) SumType {
	members := make([]SumTypeMember, len(node.Members))
	typing := map[string]ExpressionType{}
	for i, member := range node.Members {
		m, ok := checkSumTypeMember(c, member)
		members[i] = m
		if ok {
			typing[m.Name.Text()] = getSumTypeMemberType(m)
		}
	}
	if len(typing) < 2 {
		c.report("At least 2 members expected", node.Loc())
	}
	return SumType{
		Members: members,
		typing:  Type{Sum{typing}},
		loc:     node.Loc(),
	}
}

func checkSumTypeMember(c *Checker, node parser.Node) (SumTypeMember, bool) {
	if typed, ok := node.(parser.TypedExpression); ok {
		if typed.Colon {
			c.report("No ':' expected", typed.Loc())
		}
		identifier, ok := checkTypeIdentifier(c, typed.Expr)

		typing := c.checkExpression(typed.Typing)
		if typing.Type().Kind() != TYPE {
			c.report("Type expected", typing.Loc())
			ok = false
		}
		return SumTypeMember{Name: identifier, Typing: typing}, ok
	}

	identifier, ok := checkTypeIdentifier(c, node)
	if !ok {
		c.report("Type identifier expected", node.Loc())
	}
	return SumTypeMember{Name: identifier}, ok
}

func getSumTypeMemberType(member SumTypeMember) ExpressionType {
	if member.Typing == nil {
		return nil
	}

	t, ok := member.Typing.Type().(Type)
	if !ok {
		return Primitive{UNKNOWN}
	}
	return t.Value
}
