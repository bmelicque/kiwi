package parser

type Node interface {
	typeCheck(*Parser)
	Loc() Loc
	Walk(cb func(Node), skip func(Node) bool)
}
type Expression interface {
	Node
	Type() ExpressionType
}

func fallback(p *Parser) Expression {
	switch p.Peek().Kind() {
	case Bang, QuestionMark, LeftBracket:
		return p.parseUnaryExpression()
	case LeftParenthesis:
		return p.parseFunctionExpression(nil)
	case LeftBrace:
		if p.allowBraceParsing {
			return p.parseBlock()
		}
	}
	return p.parseToken()
}
