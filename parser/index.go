package parser

import "io"

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
	case AsyncKeyword, AwaitKeyword, Bang, BinaryAnd, Mul, QuestionMark, LeftBracket:
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

func Parse(reader io.Reader) ([]Node, []ParserError) {
	p := MakeParser(reader)
	statements := []Node{}

	for p.Peek().Kind() != EOF {
		statements = append(statements, p.parseStatement())
		next := p.Peek().Kind()
		if next == EOL {
			p.DiscardLineBreaks()
		} else if next != EOF {
			p.report("End of line expected", p.Peek().Loc())
		}
	}

	for i := range statements {
		statements[i].typeCheck(p)
	}

	if len(p.errors) > 0 {
		statements = []Node{}
	}
	return statements, p.errors
}
