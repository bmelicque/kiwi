package parser

type TokenExpression struct {
	Token
}

func (t TokenExpression) Loc() Loc { return t.Token.Loc() }

func (p *Parser) parseTokenExpression() Node {
	token := p.Peek()
	switch token.Kind() {
	case BooleanLiteral, NumberLiteral, StringLiteral, Name, BooleanKeyword, NumberKeyword, StringKeyword:
		p.Consume()
		return TokenExpression{token}
	}
	if !p.allowEmptyExpr {
		p.Consume()
		p.report("Expression expected", token.Loc())
	}
	return nil
}
