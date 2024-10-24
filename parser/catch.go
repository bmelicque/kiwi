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

func (c *CatchExpression) Walk(cb func(Node), skip func(Node) bool) {
	if skip(c) {
		return
	}
	cb(c)
	c.Left.Walk(cb, skip)
	if c.Body != nil {
		c.Body.Walk(cb, skip)
	}
}

func (c *CatchExpression) typeCheck(p *Parser) {
	c.Left.typeCheck(p)
	p.pushScope(NewScope(BlockScope))
	defer p.dropScope()

	var happy, err ExpressionType
	alias, ok := c.Left.Type().(TypeAlias)
	if !ok || alias.Name != "Result" {
		p.report("Catch not need (lhs is not a result type)", c.Loc())
		happy = c.Left.Type()
		err = Primitive{UNKNOWN}
	} else {
		happy = alias.Ref.(Sum).getMember("Ok")
		err = alias.Ref.(Sum).getMember("Err")
	}
	if c.Identifier != nil {
		p.scope.Add(c.Identifier.Text(), c.Identifier.Loc(), err)
	}

	c.Body.typeCheck(p)
	if !happy.Extends(c.Body.Type()) {
		p.report("Type doesn't match result's ok type", c.Body.reportLoc())
	}
}

func (c *CatchExpression) Type() ExpressionType {
	alias, ok := c.Left.Type().(TypeAlias)
	if !ok || alias.Name != "Result" {
		return c.Left.Type()
	}
	return alias.Ref.(Sum).getMember("Ok")
}

func (p *Parser) parseCatchExpression() Expression {
	expr := p.parseAccessExpression()
	if p.Peek().Kind() != CatchKeyword {
		return expr
	}
	keyword := p.Consume()

	outer := p.allowBraceParsing
	p.allowBraceParsing = false
	token := p.parseToken()
	p.allowBraceParsing = outer
	identifier, ok := token.(*Identifier)
	if token != nil && !ok {
		p.report("Identifier expected", token.Loc())
	}

	if p.Peek().Kind() != LeftBrace {
		recover(p, LeftBrace)
	}
	body := p.parseBlock()
	return &CatchExpression{
		Left:       expr,
		Keyword:    keyword,
		Identifier: identifier,
		Body:       body,
	}
}
