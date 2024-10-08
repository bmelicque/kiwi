package parser

import (
	"slices"
)

// Expr[Property]
type ComputedAccessExpression struct {
	Expr     Node
	Property BracketedExpression
}

func (c ComputedAccessExpression) Loc() Loc {
	return Loc{
		Start: c.Expr.Loc().Start,
		End:   c.Property.loc.End,
	}
}

// Callee(...Args)
type CallExpression struct {
	Callee Node
	Args   ParenthesizedExpression
}

func (c CallExpression) Loc() Loc {
	return Loc{
		Start: c.Callee.Loc().Start,
		End:   c.Args.loc.End,
	}
}

// Expr.Property
type PropertyAccessExpression struct {
	Expr     Node
	Property Node
}

func (p PropertyAccessExpression) Loc() Loc {
	return Loc{
		Start: p.Expr.Loc().Start,
		End:   p.Property.Loc().End,
	}
}

var operators = []TokenKind{LeftBracket, LeftParenthesis, Dot, LeftBrace}

func (p *Parser) parseAccessExpression() Node {
	expression := fallback(p)
	for slices.Contains(operators, p.Peek().Kind()) {
		next := p.Peek().Kind()
		isForbidden := next == LeftBrace && !p.allowBraceParsing ||
			next == LeftParenthesis && !p.allowCallExpr
		if isForbidden {
			return expression
		}
		expression = parseOneAccess(p, expression)
	}
	return expression
}

func parseOneAccess(p *Parser, expr Node) Node {
	next := p.Peek()
	switch next.Kind() {
	case LeftBracket:
		return ComputedAccessExpression{expr, p.parseBracketedExpression()}
	case LeftParenthesis:
		return CallExpression{expr, p.parseParenthesizedExpression()}
	case Dot:
		p.Consume()
		tmp := p.allowCallExpr
		p.allowCallExpr = false
		property := fallback(p)
		p.allowCallExpr = tmp
		return PropertyAccessExpression{
			Expr:     expr,
			Property: property,
		}
	default:
		panic("switch should've been exhaustive!")
	}
}
