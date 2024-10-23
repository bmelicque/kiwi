package parser

type IfExpression struct {
	Keyword   Token
	Condition Node       // *Assignment | Expression
	Alternate Expression // *IfExpression | *Block
	Body      *Block
}

func (i *IfExpression) Walk(cb func(Node), skip func(Node) bool) {
	if skip(i) {
		return
	}
	cb(i)
	if i.Condition != nil {
		i.Condition.Walk(cb, skip)
	}
	if i.Body != nil {
		i.Body.Walk(cb, skip)
	}
	if i.Alternate != nil {
		i.Alternate.Walk(cb, skip)
	}
}

func (i *IfExpression) typeCheck(p *Parser) {
	p.pushScope(NewScope(BlockScope))

	outer := p.conditionalDeclaration
	p.conditionalDeclaration = true
	i.Condition.typeCheck(p)
	p.conditionalDeclaration = outer

	if expr, ok := i.Condition.(Expression); ok {
		if expr.Type().Kind() != BOOLEAN {
			p.report("Expected boolean condition", i.Condition.Loc())
		}
	}
	i.Body.typeCheck(p)
	p.dropScope()

	if i.Alternate == nil {
		return
	}
	i.Alternate.typeCheck(p)
	if !Match(i.Body.Type(), i.Alternate.Type()) {
		loc := Loc{i.Keyword.Loc().Start, i.Alternate.Loc().End}
		p.report("Types of the main and alternate blocks don't match", loc)
	}
}

func (i *IfExpression) Loc() Loc {
	return Loc{
		Start: i.Keyword.Loc().Start,
		End:   i.Body.Loc().End,
	}
}
func (i *IfExpression) Type() ExpressionType {
	if i.Alternate == nil {
		return makeOptionType(i.Body.Type())
	}
	return i.Alternate.Type()
}

func (p *Parser) parseIfExpression() *IfExpression {
	keyword := p.Consume() // "if" keyword
	condition := parseIfCondition(p)
	body := parseIfBody(p)
	alternate := parseAlternate(p)
	return &IfExpression{keyword, condition, alternate, body}
}

// Parse the condition of an If expression: if condition {...}
func parseIfCondition(p *Parser) Node {
	outer := p.allowBraceParsing
	p.allowBraceParsing = false
	condition := p.parseAssignment()
	p.allowBraceParsing = outer
	return condition
}

// Parse the body of an If expression
func parseIfBody(p *Parser) *Block {
	if p.Peek().Kind() != LeftBrace {
		p.report("Block expected", p.Peek().Loc())
		return nil
	}
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
