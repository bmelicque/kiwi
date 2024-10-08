package parser

type TokenExpression struct {
	Token
}

func (t TokenExpression) Loc() Loc { return t.Token.Loc() }

func (p *Parser) parseTokenExpression() Node {
	token := p.Peek()
	switch token.Kind() {
	case BOOLEAN, NUMBER, STRING, IDENTIFIER, BOOL_KW, NUM_KW, STR_KW:
		p.Consume()
		return TokenExpression{token}
	}
	if !p.allowEmptyExpr {
		p.Consume()
		p.report("Expression expected", token.Loc())
	}
	return nil
}
