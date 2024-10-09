package parser

type Node interface {
	Loc() Loc
}
type Expression interface {
	Node
	Type() ExpressionType
}

func fallback(p *Parser) Expression {
	switch p.Peek().Kind() {
	case LeftBracket:
		return p.parseUnaryExpression()
	case LeftParenthesis:
		return p.parseFunctionExpression(nil)
	case LeftBrace:
		if p.allowBraceParsing {
			return p.parseBlock()
		}
	}
	return p.parseToken(false)
}
