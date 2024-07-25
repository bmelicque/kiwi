package parser

import (
	"fmt"
	"slices"

	"github.com/bmelicque/test-parser/tokenizer"
)

// TODO: random access expression

type CallExpression struct {
	Callee Expression
	Args   Expression
}

func (c CallExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: c.Callee.Loc().Start,
		End:   c.Args.Loc().End,
	}
}
func (c CallExpression) Type(ctx *Scope) ExpressionType {
	callee := c.Callee
	if callee == nil {
		return nil
	}

	if calleeType, ok := callee.Type(ctx).(Function); ok {
		return calleeType.returned
	} else {
		return nil
	}
}
func (expr CallExpression) Check(c *Checker) {
	expr.Callee.Check(c)
	expr.Args.Check(c)

	if _, ok := expr.Args.(TupleExpression); !ok {
		c.report("Tuple expression expected", expr.Args.Loc())
		return
	}

	calleeType, ok := expr.Callee.Type(c.scope).(Function)
	if !ok {
		c.report("Function type expected", expr.Callee.Loc())
		return
	}

	if !(expr.Args.Type(c.scope).Extends(calleeType.params)) {
		c.report("Arguments types don't match expected parameters types", expr.Args.Loc())
	}
}

type PropertyAccessExpression struct {
	Expr     Expression
	Property Expression
}

func (p PropertyAccessExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: p.Expr.Loc().Start,
		End:   p.Property.Loc().End,
	}
}
func (p PropertyAccessExpression) Type(ctx *Scope) ExpressionType {
	token, ok := p.Property.(TokenExpression)
	if !ok {
		return nil
	}
	prop := token.Token.Text()

	typing := p.Expr.Type(ctx)
	if ref, ok := typing.(TypeRef); ok {
		typing = ref.ref
	}
	if struc, ok := typing.(Object); ok {
		return struc.members[prop]
	} else {
		return nil
	}
}
func (p PropertyAccessExpression) Check(c *Checker) {
	p.Expr.Check(c)
	typing := p.Expr.Type(c.scope)
	if ref, ok := typing.(TypeRef); ok {
		typing = ref.ref
	}
	struc, ok := typing.(Object)
	if !ok {
		c.report("Object type expected", p.Expr.Loc())
	}

	switch prop := p.Property.(type) {
	case TokenExpression:
		if ok {
			name := prop.Token.Text()
			_, ok := struc.members[name]
			if !ok {
				c.report(fmt.Sprintf("Property '%v' does not exist on this type", name), prop.Loc())
			}
		}
	default:
		c.report("Identifier expected", prop.Loc())
	}
}

type ObjectExpression struct {
	typing  Expression
	members []Expression
	loc     tokenizer.Loc
}

func (o ObjectExpression) Loc() tokenizer.Loc             { return o.loc }
func (o ObjectExpression) Type(ctx *Scope) ExpressionType { return o.typing.Type(ctx) }
func (o ObjectExpression) Check(c *Checker) {
	typing := getValidatedObjectType(c, o)

	members := map[string]bool{}
	for _, member := range o.members {
		member, ok := member.(TypedExpression)
		if !ok {
			c.report("Member expression expected", member.Loc())
			continue
		}
		name := getValidatedMemberName(c, member)
		if name != "" {
			members[name] = true
		}
		checkMemberValue(c, member, typing.members[name])
	}

	if typing != nil {
		for name := range typing.members {
			if _, ok := members[name]; !ok {
				c.report(fmt.Sprintf("Missing key '%v'", name), o.Loc())
			}
		}
	}
}
func getValidatedObjectType(c *Checker, s ObjectExpression) *Object {
	if s.typing == nil {
		return nil
	}
	s.typing.Check(c)
	typing := s.typing.Type(c.scope).(Type)
	if t, ok := typing.value.(Object); ok {
		return &t
	}
	c.report("Object type expected", s.typing.Loc())
	return nil
}
func getValidatedMemberName(c *Checker, member TypedExpression) string {
	expr, ok := member.Expr.(TokenExpression)
	if !ok || expr.Token.Kind() != tokenizer.IDENTIFIER {
		c.report("Identifier expected", member.Expr.Loc())
		return ""
	}
	return expr.Token.Text()
}
func checkMemberValue(c *Checker, member TypedExpression, expected ExpressionType) {
	if member.typing == nil {
		c.report("Value expected", member.Loc())
		return
	}
	member.typing.Check(c)
	if expected == nil {
		c.report("Property does not exist on type", member.Expr.Loc())
		return
	}
	if !expected.Extends(member.typing.Type(c.scope)) {
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
			expression = PropertyAccessExpression{expression, property}
		case tokenizer.LBRACE:
			if !IsType(expression) {
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
				typing:  expression,
				members: members,
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
