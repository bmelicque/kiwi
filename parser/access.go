package parser

import (
	"slices"

	"github.com/bmelicque/test-parser/tokenizer"
)

// Expr[Property]
type ComputedAccessExpression struct {
	Expr     Node
	Property BracketedExpression
}

func (c ComputedAccessExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: c.Expr.Loc().Start,
		End:   c.Property.loc.End,
	}
}

// Callee(...Args)
type CallExpression struct {
	Callee Node
	Args   ParenthesizedExpression
}

func (c CallExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: c.Callee.Loc().Start,
		End:   c.Args.loc.End,
	}
}

// Expr.Property
type PropertyAccessExpression struct {
	Expr     Node
	Property Node
}

func (p PropertyAccessExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: p.Expr.Loc().Start,
		End:   p.Property.Loc().End,
	}
}

// Typing{...Members}
type InstanciationExpression struct {
	Typing  Node
	Members []Node
	loc     tokenizer.Loc
}

func (i InstanciationExpression) Loc() tokenizer.Loc { return i.loc }

var operators = []tokenizer.TokenKind{tokenizer.LBRACKET, tokenizer.LPAREN, tokenizer.DOT, tokenizer.LBRACE}

func (p *Parser) parseAccessExpression() Node {
	expression := fallback(p)
	for slices.Contains(operators, p.tokenizer.Peek().Kind()) {
		next := p.tokenizer.Peek().Kind()
		isForbidden := next == tokenizer.LBRACE && !p.allowBraceParsing ||
			next == tokenizer.LPAREN && !p.allowCallExpr
		if isForbidden {
			return expression
		}
		expression = parseOneAccess(p, expression)
	}
	return expression
}

func parseOneAccess(p *Parser, expr Node) Node {
	next := p.tokenizer.Peek()
	switch next.Kind() {
	case tokenizer.LBRACKET:
		return ComputedAccessExpression{expr, p.parseBracketedExpression()}
	case tokenizer.LPAREN:
		return CallExpression{expr, p.parseParenthesizedExpression()}
	case tokenizer.DOT:
		p.tokenizer.Consume()
		tmp := p.allowCallExpr
		p.allowCallExpr = false
		property := fallback(p)
		p.allowCallExpr = tmp
		return PropertyAccessExpression{
			Expr:     expr,
			Property: property,
		}
	case tokenizer.LBRACE:
		if !p.allowBraceParsing {
			return expr
		}
		// TODO: parseTuple
		p.tokenizer.Consume()
		var members []Node
		ParseList(p, tokenizer.RBRACE, func() {
			members = append(members, p.parseTypedExpression())
		})
		loc := tokenizer.Loc{Start: expr.Loc().Start}
		if p.tokenizer.Peek().Kind() != tokenizer.RBRACE {
			p.report("'}' expected", p.tokenizer.Peek().Loc())
		} else {
			loc.End = p.tokenizer.Consume().Loc().End
		}
		return InstanciationExpression{
			Typing:  expr,
			Members: members,
			loc:     loc,
		}
	default:
		panic("switch should've been exhaustive!")
	}
}
