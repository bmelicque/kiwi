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
	if struc, ok := typing.(Struct); ok {
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
	struc, ok := typing.(Struct)
	if !ok {
		c.report("Structured type expected", p.Expr.Loc())
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

var operators = []tokenizer.TokenKind{tokenizer.LPAREN, tokenizer.DOT}

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
		return ParseStructDef(p)
	}
	return TokenExpression{}.Parse(p)
	// TODO: if IsType, check for { }
}
