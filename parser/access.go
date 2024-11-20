package parser

import (
	"slices"
)

var operators = []TokenKind{LeftBracket, LeftParenthesis, Dot}

func (p *Parser) parseAccessExpression() Expression {
	expression := fallback(p)
	for slices.Contains(operators, p.Peek().Kind()) {
		next := p.Peek().Kind()
		if next == LeftParenthesis && !p.allowCallExpr {
			return expression
		}
		expression = parseOneAccess(p, expression)
	}
	return expression
}

func parseOneAccess(p *Parser, expr Expression) Expression {
	next := p.Peek()
	switch next.Kind() {
	case LeftBracket:
		return parseComputedAccessExpression(p, expr)
	case LeftParenthesis:
		return parseCallExpression(p, expr)
	case Dot:
		return parsePropertyAccess(p, expr)
	default:
		panic("switch should've been exhaustive!")
	}
}
