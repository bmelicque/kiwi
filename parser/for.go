package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type For struct {
	Keyword   tokenizer.Token
	Statement Node // ExpressionStatement holding a condition OR Assignment
	Body      *Block
}

func (f For) Loc() tokenizer.Loc {
	loc := f.Keyword.Loc()
	if f.Body != nil {
		loc.End = f.Body.Loc().End
	} else if f.Statement != nil {
		loc.End = f.Statement.Loc().End
	}
	return loc
}

func ParseForLoop(p *Parser) Node {
	statement := For{}
	statement.Keyword = p.tokenizer.Consume()

	outer := p.allowBraceParsing
	p.allowBraceParsing = false
	statement.Statement = p.parseAssignment()
	p.allowBraceParsing = outer
	statement.Body = p.parseBlock()

	return statement
}
