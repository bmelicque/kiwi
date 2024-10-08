package parser

type ForExpression struct {
	Keyword   Token
	Statement Node // ExpressionStatement holding a condition OR Assignment
	Body      *Block
}

func (f ForExpression) Loc() Loc {
	loc := f.Keyword.Loc()
	if f.Body != nil {
		loc.End = f.Body.Loc().End
	} else if f.Statement != nil {
		loc.End = f.Statement.Loc().End
	}
	return loc
}

func (p *Parser) parseForExpression() ForExpression {
	keyword := p.Consume()
	outer := p.allowBraceParsing
	p.allowBraceParsing = false
	statement := p.parseAssignment()
	p.allowBraceParsing = outer
	block := p.parseBlock()
	return ForExpression{keyword, statement, block}
}
