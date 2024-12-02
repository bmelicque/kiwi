package parser

type IfExpression struct {
	Keyword   Token
	Condition Node       // *Assignment | Expression
	Alternate Expression // *IfExpression | *Block
	Body      *Block
}

func (i *IfExpression) getChildren() []Node {
	children := make([]Node, 0, 3)
	if i.Condition != nil {
		children = append(children, i.Condition)
	}
	if i.Body != nil {
		children = append(children, i.Body)
	}
	if i.Alternate != nil {
		children = append(children, i.Alternate)
	}
	return children
}

func (i *IfExpression) typeCheck(p *Parser) {
	p.pushScope(NewScope(BlockScope))

	outer := p.conditionalDeclaration
	p.conditionalDeclaration = true
	i.Condition.typeCheck(p)
	p.conditionalDeclaration = outer

	if expr, ok := i.Condition.(Expression); ok {
		if _, ok := expr.Type().(Boolean); !ok {
			p.error(i.Condition, BooleanExpected, expr.Type())
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
		p.error(&Block{loc: loc}, CannotAssignType, i.Body.Type(), i.Alternate.Type())
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
		p.error(&Literal{p.Peek()}, TokenExpected, LeftBrace)
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
		p.error(&Literal{p.Peek()}, TokenExpected, LeftBrace)
		return nil
	}
}
