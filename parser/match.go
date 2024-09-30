package parser

import (
	"slices"

	"github.com/bmelicque/test-parser/tokenizer"
)

type MatchCase struct {
	Pattern    Node
	Statements []Node
}

type MatchStatement struct {
	Keyword tokenizer.Token
	Value   Node
	Cases   []MatchCase
	end     tokenizer.Position
}

func (m MatchStatement) Loc() tokenizer.Loc {
	loc := m.Keyword.Loc()
	if m.end != (tokenizer.Position{}) {
		loc.End = m.end
	}
	return loc
}

func (p *Parser) parseMatchStatement() Node {
	keyword := p.tokenizer.Consume()
	outer := p.allowBraceParsing
	p.allowBraceParsing = false
	condition := ParseExpression(p)
	p.allowBraceParsing = outer
	if p.tokenizer.Peek().Kind() != tokenizer.LBRACE && !recover(p, tokenizer.LBRACE) {
		return MatchStatement{Keyword: keyword, Value: condition}
	}
	p.tokenizer.Consume()
	p.tokenizer.DiscardLineBreaks()

	cases := []MatchCase{}
	stopAt := []tokenizer.TokenKind{tokenizer.RBRACE, tokenizer.EOF}
	for !slices.Contains(stopAt, p.tokenizer.Peek().Kind()) {
		cases = append(cases, parseMatchCase(p))
	}

	next := p.tokenizer.Peek()
	end := next.Loc().End
	if next.Kind() == tokenizer.RBRACE {
		p.tokenizer.Consume()
	} else {
		p.report("'}' expected", next.Loc())
	}
	return MatchStatement{keyword, condition, cases, end}
}

func parseMatchCase(p *Parser) MatchCase {
	var pattern Node
	if p.tokenizer.Peek().Kind() == tokenizer.CASE_KW {
		p.tokenizer.Consume()
		pattern = ParseExpression(p)
		if p.tokenizer.Peek().Kind() == tokenizer.COLON {
			p.tokenizer.Consume()
		} else {
			p.report("':' expected", p.tokenizer.Peek().Loc())
		}
	}

	stopAt := []tokenizer.TokenKind{tokenizer.EOF, tokenizer.RBRACE, tokenizer.CASE_KW}
	statements := []Node{}
	for !slices.Contains(stopAt, p.tokenizer.Peek().Kind()) {
		statements = append(statements, p.parseStatement())
		p.tokenizer.DiscardLineBreaks()
	}

	return MatchCase{
		Pattern:    pattern,
		Statements: statements,
	}
}
