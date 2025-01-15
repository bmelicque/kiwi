package parser

type CatchExpression struct {
	Left       Expression
	Keyword    Token
	Identifier *Identifier
	Body       *Block
}

func (c *CatchExpression) Loc() Loc {
	start := c.Left.Loc().Start
	var end Position
	if c.Body != nil {
		end = c.Body.loc.End
	} else if c.Identifier != nil {
		end = c.Identifier.Loc().End
	} else {
		end = c.Keyword.Loc().End
	}
	return Loc{start, end}
}

func (c *CatchExpression) getChildren() []Node {
	children := []Node{c.Left}
	if c.Body != nil {
		children = append(children, c.Body)
	}
	return children
}

func (c *CatchExpression) typeCheck(p *Parser) {
	c.Left.typeCheck(p)
	p.pushScope(NewScope(BlockScope))
	defer p.dropScope()
	happy, err, ok := getCatchTypes(c.Left)
	if !ok {
		p.error(c, UnneededCatch)
	}
	if c.Identifier != nil {
		p.scope.Add(c.Identifier.Text(), c.Identifier.Loc(), err)
	}
	c.typeCheckBody(p, happy)
}

// returns (Left, Right, ok), with CatchExpression being:
// Left catch (identifier Right) {}
func getCatchTypes(result Expression) (ExpressionType, ExpressionType, bool) {
	alias, ok := result.Type().(TypeAlias)
	if !ok || alias.Name != "!" {
		return result.Type(), Unknown{}, false
	}
	happy := alias.Ref.(Sum).getMember("Ok")
	err := alias.Ref.(Sum).getMember("Err")
	return happy, err, true
}
func (c *CatchExpression) typeCheckBody(p *Parser, happy ExpressionType) {
	if c.Body == nil {
		return
	}
	c.Body.typeCheck(p)
	if !isExiting(c.Body) && !happy.Extends(c.Body.Type()) {
		p.error(c.Body.reportedNode(), CannotAssignType, happy, c.Body.Type())
	}
}

func (c *CatchExpression) Type() ExpressionType {
	t, _, _ := getCatchTypes(c.Left)
	return t
}

func (p *Parser) parseCatchExpression() Expression {
	var expr Expression
	if p.allowBraceParsing {
		expr = p.parseInstanceExpression()
	} else {
		expr = p.parseUnaryExpression()
	}
	if p.Peek().Kind() != CatchKeyword {
		return expr
	}
	keyword := p.Consume()

	identifier := parseCatchIdentifier(p)
	body := parseCatchBody(p)

	return &CatchExpression{
		Left:       expr,
		Keyword:    keyword,
		Identifier: identifier,
		Body:       body,
	}
}

func parseCatchIdentifier(p *Parser) *Identifier {
	outerBrace, outerEmpty := p.allowBraceParsing, p.allowEmptyExpr
	p.allowBraceParsing, p.allowEmptyExpr = false, true
	token := p.parseToken()
	p.allowBraceParsing, p.allowEmptyExpr = outerBrace, outerEmpty
	identifier, ok := token.(*Identifier)
	if token != nil && !ok {
		p.error(token, IdentifierExpected)
	}
	return identifier
}

func parseCatchBody(p *Parser) *Block {
	if p.Peek().Kind() != LeftBrace {
		recoverBadTokens(p, LeftBrace)
	}
	// left brace if recovered correctly, else EOL/EOF
	if p.Peek().Kind() != LeftBrace {
		return nil
	}
	return p.parseBlock()
}
