package parser

type Node interface {
	Loc() Loc
}

func fallback(p *Parser) Node {
	switch p.Peek().Kind() {
	case LeftBracket:
		return p.parseUnaryExpression()
	case LeftParenthesis:
		return p.parseFunctionExpression()
	case LeftBrace:
		if p.allowBraceParsing {
			return p.parseBlock()
		}
	}
	return p.parseTokenExpression()
}
