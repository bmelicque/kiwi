package parser

type Node interface {
	typeCheck(*Parser)
	Loc() Loc
	getChildren() []Node
}
type Expression interface {
	Node
	Type() ExpressionType
}

func fallback(p *Parser) Expression {
	switch p.Peek().Kind() {
	case AsyncKeyword:
		return p.parseAsyncExpression()
	case AwaitKeyword:
		return p.parseAwaitExpression()
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

func Walk(node Node, predicate func(n Node, skip func())) {
	var s bool
	predicate(node, func() { s = true })
	if s {
		return
	}
	children := node.getChildren()
	for i := range children {
		Walk(children[i], predicate)
	}
}
