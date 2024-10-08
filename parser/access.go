package parser

import (
	"slices"
)

var operators = []TokenKind{LeftBracket, LeftParenthesis, Dot, LeftBrace}

func (p *Parser) parseAccessExpression() Expression {
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

func parseOneAccess(p *Parser, expr Expression) Expression {
	next := p.Peek()
	switch next.Kind() {
	case LeftBracket:
		return parseComputedAccessExpression(p, expr)
	case LeftParenthesis:
		_, isType := expr.Type().(Type)
		if isType {
			return parseInstanciation(p, expr)
		} else {
			return parseFunctionCall(p, expr)
		}
	case Dot:
		return parsePropertyAccess(p, expr)
	default:
		panic("switch should've been exhaustive!")
	}
}
