package parser

type IfExpression struct {
	Keyword   Token
	Condition Expression
	Alternate Expression // *IfExpression | *Block
	Body      *Block
}

func (i IfExpression) Loc() Loc {
	return Loc{
		Start: i.Keyword.Loc().Start,
		End:   i.Body.Loc().End,
	}
}
func (i IfExpression) Type() ExpressionType {
	if i.Alternate == nil {
		return makeOptionType(i.Alternate.Type())
	}
	return i.Alternate.Type()
}

func (p *Parser) parseIfExpression() *IfExpression {
	keyword := p.Consume() // "if" keyword
	condition := parseIfCondition(p)
	body := parseIfBody(p)
	alternate := parseAlternate(p)
	if alternate != nil && !Match(body.Type(), alternate.Type()) {
		loc := Loc{keyword.Loc().Start, alternate.Loc().End}
		p.report("Types of the main and alternate blocks don't match", loc)
	}
	return &IfExpression{keyword, condition, alternate, body}
}

// Parse the condition of an If expression: if condition {...}
func parseIfCondition(p *Parser) Expression {
	outer := p.allowBraceParsing
	p.allowBraceParsing = false
	condition := p.parseExpression()
	p.allowBraceParsing = outer
	if condition.Type().Kind() != BOOLEAN {
		p.report("Expected boolean condition", condition.Loc())
	}
	return condition
}

// Parse the body of an If expression
func parseIfBody(p *Parser) *Block {
	if p.Peek().Kind() != LeftBrace {
		p.report("Block expected", p.Peek().Loc())
		return nil
	}
	p.pushScope(NewScope(BlockScope))
	defer p.dropScope()
	return p.parseBlock()
}

// Parse a potential alternate for an If expression.
// Candidates are a block or another If expression.
func parseAlternate(p *Parser) Expression {
	if p.Peek().Kind() != ElseKeyword {
		return nil
	}
	p.Consume() // "else"
	switch p.Peek().Kind() {
	case IfKeyword:
		return p.parseIfExpression()
	case LeftBrace:
		return p.parseBlock()
	default:
		p.report("Block expected", p.Peek().Loc())
		return nil
	}
}
