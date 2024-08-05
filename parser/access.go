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
	loc := tokenizer.Loc{}
	if c.Callee != nil {
		loc.Start = c.Callee.Loc().Start
	} else {
		loc.Start = c.Args.Loc().Start
	}
	if c.Args != nil {
		loc.End = c.Args.Loc().End
	} else {
		loc.End = c.Callee.Loc().End
	}
	return loc
}
func (c CallExpression) Type() ExpressionType {
	callee := c.Callee
	if callee == nil {
		return nil
	}

	if calleeType, ok := callee.Type().(Function); ok {
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

	calleeType, ok := expr.Callee.Type().(Function)
	if !ok {
		c.report("Function type expected", expr.Callee.Loc())
		return
	}

	if !(expr.Args.Type().Extends(calleeType.params)) {
		c.report("Arguments types don't match expected parameters types", expr.Args.Loc())
	}
}

type PropertyAccessExpression struct {
	Expr     Expression
	Property Expression
	typing   ExpressionType
	method   bool
}

func (p PropertyAccessExpression) IsMethod() bool { return p.method }

func (p *PropertyAccessExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: p.Expr.Loc().Start,
		End:   p.Property.Loc().End,
	}
}
func (p *PropertyAccessExpression) setType(ctx *Scope) {
	token, ok := p.Property.(*TokenExpression)
	if !ok {
		return
	}
	prop := token.Token.Text()

	typing := p.Expr.Type()
	ref, _ := typing.(TypeRef)
	object, ok := ref.ref.(Object)
	if !ok {
		return
	}
	if method, ok := ctx.FindMethod(prop, typing); ok {
		p.method = true
		p.typing = method.signature
	} else {
		p.typing = object.members[prop]
	}

}
func (p *PropertyAccessExpression) Type() ExpressionType { return p.typing }
func (p *PropertyAccessExpression) Check(c *Checker) {
	p.Expr.Check(c)

	switch prop := p.Property.(type) {
	case *TokenExpression:
		p.setType(c.scope)
		typing := p.Expr.Type()
		ref, _ := typing.(TypeRef)
		if _, ok := ref.ref.(Object); !ok {
			return
		}
		if p.typing == nil {
			c.report(fmt.Sprintf("Property '%v' does not exist on this type", prop.Token.Text()), prop.Loc())
		}
	default:
		c.report("Identifier expected", prop.Loc())
	}
}

type ObjectExpression struct {
	typing  Expression
	Members []Expression
	loc     tokenizer.Loc
}

func (o ObjectExpression) Loc() tokenizer.Loc   { return o.loc }
func (o ObjectExpression) Type() ExpressionType { return o.typing.Type() }
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
			checkMemberValue(c, member, typing.members[name])
		}
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
	typing, _ := s.typing.Type().(TypeRef)
	object, ok := typing.ref.(Object)
	if !ok {
		c.report("Object type expected", s.typing.Loc())
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
