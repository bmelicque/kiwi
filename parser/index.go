package parser

type Node interface {
	Loc() Loc
}

func fallback(p *Parser) Node {
	switch p.Peek().Kind() {
	case LBRACKET:
		return p.parseUnaryExpression()
	case LPAREN:
		return p.parseFunctionExpression()
	case LBRACE:
		if p.allowBraceParsing {
			return p.parseBlock()
		}
	}
	return p.parseTokenExpression()
}
