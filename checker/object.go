package checker

import (
	"fmt"
	"slices"

	"github.com/bmelicque/test-parser/tokenizer"
)

type ObjectExpression struct {
	Typing  Expression
	Members []Expression
	loc     tokenizer.Loc
}

func (o ObjectExpression) Loc() tokenizer.Loc   { return o.loc }
func (o ObjectExpression) Type() ExpressionType { return o.Typing.Type() }
func (o ObjectExpression) Check(c *Checker) {
	typing := getValidatedObjectType(c, o)

	members := map[string]bool{}
	for _, member := range o.Members {
		member, ok := member.(TypedExpression)
		if !ok {
			c.report("Member expression expected", member.Loc())
			continue
		}
		name := getValidatedMemberName(c, member)
		if name != "" {
			members[name] = true
		}
		if typing != nil {
			checkMemberValue(c, member, typing.Members[name])
		}
	}

	if typing != nil {
		for name := range typing.Members {
			if _, ok := members[name]; !ok {
				c.report(fmt.Sprintf("Missing key '%v'", name), o.Loc())
			}
		}
	}
}
func getValidatedObjectType(c *Checker, s ObjectExpression) *Object {
	if s.Typing == nil {
		return nil
	}
	s.Typing.Check(c)
	typing, _ := s.Typing.Type().(TypeRef)
	object, ok := typing.Ref.(Object)
	if !ok {
		c.report("Object type expected", s.Typing.Loc())
		return nil
	}
	return &object

}
func getValidatedMemberName(c *Checker, member TypedExpression) string {
	expr, ok := member.Expr.(*TokenExpression)
	if !ok || expr.Token.Kind() != tokenizer.IDENTIFIER {
		c.report("Identifier expected", member.Expr.Loc())
		return ""
	}
	return expr.Token.Text()
}
func checkMemberValue(c *Checker, member TypedExpression, expected ExpressionType) {
	if member.Typing == nil {
		c.report("Value expected", member.Loc())
		return
	}
	member.Typing.Check(c)
	if expected == nil {
		c.report("Property does not exist on type", member.Expr.Loc())
		return
	}
	if !expected.Extends(member.Typing.Type()) {
		c.report("Types do not match", member.Loc())
	}
}

var operators = []tokenizer.TokenKind{tokenizer.LPAREN, tokenizer.DOT, tokenizer.LBRACE}

func ParseAccessExpression(p *Parser) Expression {
	expression := fallback(p)
	next := p.tokenizer.Peek()
	for slices.Contains(operators, next.Kind()) {
		switch next.Kind() {
		case tokenizer.LPAREN:
			args := ParseTupleExpression(p)
			expression = CallExpression{expression, args}
		case tokenizer.DOT:
			p.tokenizer.Consume()
			property := fallback(p)
			expression = &PropertyAccessExpression{
				Expr:     expression,
				Property: property,
			}
		case tokenizer.LBRACE:
			if !IsTypeToken(expression) {
				return expression
			}
			p.tokenizer.Consume()
			members := []Expression{}
			ParseList(p, tokenizer.RBRACE, func() {
				members = append(members, ParseTypedExpression(p))
			})
			loc := tokenizer.Loc{
				Start: expression.Loc().Start,
				End:   members[len(members)-1].Loc().End,
			}
			if p.tokenizer.Peek().Kind() != tokenizer.RBRACE {
				p.report("'}' expected", p.tokenizer.Peek().Loc())
			} else {
				loc.End = p.tokenizer.Consume().Loc().End
			}
			return ObjectExpression{
				Typing:  expression,
				Members: members,
				loc:     loc,
			}
		}
		next = p.tokenizer.Peek()
	}
	return expression
}

func fallback(p *Parser) Expression {
	switch p.tokenizer.Peek().Kind() {
	case tokenizer.LPAREN:
		return ParseFunctionExpression(p)
	case tokenizer.LBRACKET:
		return ListExpression{}.Parse(p)
	case tokenizer.LBRACE:
		return ParseObjectDefinition(p)
	}
	return TokenExpression{}.Parse(p)
}
