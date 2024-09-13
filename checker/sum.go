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
	typing  Sum
	loc     tokenizer.Loc
}

func (s SumType) Loc() tokenizer.Loc   { return s.loc }
func (s SumType) Type() ExpressionType { return s.typing }

func (c *Checker) checkSumType(node parser.SumType) SumType {
	loc := node.Loc()
	members := make([]SumTypeMember, len(node.Members))
	typing := map[string]ExpressionType{}
	for i, member := range node.Members {
		m, ok := checkSumTypeMember(c, member)
		members[i] = m
		if !ok {
			continue
		}
		loc.End = member.Loc().End
		if m.Typing != nil {
			typing[m.Name.Text()] = m.Typing.Type()
		} else {
			typing[m.Name.Text()] = nil
		}
	}
	return SumType{
		Members: members,
		typing:  Sum{typing},
		loc:     loc,
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
	return SumTypeMember{Name: identifier}, ok
}

func checkTypeIdentifier(c *Checker, node parser.Node) (Identifier, bool) {
	token, ok := node.(parser.TokenExpression)
	if !ok {
		c.report("Identifier expected", node.Loc())
		return Identifier{}, false
	}

	identifier, ok := c.checkToken(token, false).(Identifier)
	if !ok {
		c.report("Identifier expected", node.Loc())
		return Identifier{}, false
	}
	if !identifier.isType {
		c.report("Pascal-case expected", node.Loc())
	}
	return identifier, ok
}
