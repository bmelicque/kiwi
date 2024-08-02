package parser

import (
	"fmt"

	"github.com/bmelicque/test-parser/tokenizer"
)

type ObjectDefinition struct {
	members []Expression // []TypedExpression
	loc     tokenizer.Loc
}

func (expr ObjectDefinition) Loc() tokenizer.Loc { return expr.loc }

func (expr ObjectDefinition) Check(c *Checker) {
	members := map[string][]tokenizer.Loc{}
	for _, member := range expr.members {
		name, ok := CheckTypedIdentifier(c, member)
		if ok {
			members[name] = append(members[name], member.Loc())
		}
	}

	for member, locs := range members {
		if len(locs) > 1 {
			for _, loc := range locs {
				c.report(fmt.Sprintf("Duplicate identifier '%v'", member), loc)
			}
		}
	}
}

func (expr ObjectDefinition) Type() ExpressionType {
	value := Object{map[string]ExpressionType{}}
	for _, member := range expr.members {
		name := member.(TypedExpression).Expr.(*TokenExpression).Token.Text()
		typing := member.Type()
		value.members[name] = typing
	}
	return Type{value}
}

func ParseObjectDefinition(p *Parser) Expression {
	lbrace := p.tokenizer.Consume()
	loc := lbrace.Loc()

	members := []Expression{}
	ParseList(p, tokenizer.RBRACE, func() {
		members = append(members, ParseTypedExpression(p))
	})

	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.RBRACE {
		p.report("'}' expected", next.Loc())
	}
	loc.End = p.tokenizer.Consume().Loc().End

	return ObjectDefinition{members, loc}
}
