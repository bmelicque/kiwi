package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type SumType struct {
	Members []Node
	start   tokenizer.Position
}

func (s SumType) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: s.start,
		End:   s.Members[len(s.Members)-1].Loc().End,
	}
}

func (p *Parser) parseSumType() Node {
	if p.tokenizer.Peek().Kind() != tokenizer.BOR {
		return p.parseTypedExpression()
	}

	start := p.tokenizer.Peek().Loc().Start
	members := []Node{}
	for p.tokenizer.Peek().Kind() == tokenizer.BOR {
		p.tokenizer.Consume()
		members = append(members, p.parseTypedExpression())
		handleSumTypeBadTokens(p)
		p.tokenizer.DiscardLineBreaks()
	}
	return SumType{Members: members, start: start}
}

func handleSumTypeBadTokens(p *Parser) {
	err := false
	var start, end tokenizer.Position
	for p.tokenizer.Peek().Kind() != tokenizer.EOL && p.tokenizer.Peek().Kind() != tokenizer.EOF && p.tokenizer.Peek().Kind() != tokenizer.BOR {
		token := p.tokenizer.Consume()
		if !err {
			err = true
			start = token.Loc().Start
		}
		end = token.Loc().End
	}
	if err {
		p.report("EOL or '|' expected", tokenizer.Loc{Start: start, End: end})
	}
}
