package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type For struct {
	Keyword   tokenizer.Token
	Statement Node // ExpressionStatement holding a condition OR Assignment
	Body      *Body
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

	if p.tokenizer.Peek().Kind() != tokenizer.LBRACE {
		statement.Statement = ParseAssignment(p)
	}

	statement.Body = ParseBody(p)

	return statement
}
