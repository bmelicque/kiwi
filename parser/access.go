package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

// TODO: member expression
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

func ParseCallExpression(p *Parser) Expression {
	expression := fallback(p)
	next := p.tokenizer.Peek()
	for next.Kind() == tokenizer.LPAREN {
		args := ParseTupleExpression(p)
		expression = CallExpression{expression, args}
		next = p.tokenizer.Peek()
	}
	return expression
}

func fallback(p *Parser) Expression {
	if p.tokenizer.Peek().Kind() == tokenizer.LPAREN {
		return ParseFunctionExpression(p)
	}
	if p.tokenizer.Peek().Kind() == tokenizer.LBRACKET {
		return ListExpression{}.Parse(p)
	}
	return TokenExpression{}.Parse(p)
}
