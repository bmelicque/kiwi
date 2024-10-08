package parser

import "slices"

type MatchCase struct {
	Pattern    Node
	Statements []Node
}

type MatchExpression struct {
	Keyword Token
	Value   Node
	Cases   []MatchCase
	end     Position
}

func (m MatchExpression) Loc() Loc {
	loc := m.Keyword.Loc()
	if m.end != (Position{}) {
		loc.End = m.end
	}
	return loc
}

func (p *Parser) parseMatchExpression() Node {
	keyword := p.Consume()
	outer := p.allowBraceParsing
	p.allowBraceParsing = false
	condition := ParseExpression(p)
	p.allowBraceParsing = outer
	if p.Peek().Kind() != LeftBrace && !recover(p, LeftBrace) {
		return MatchExpression{Keyword: keyword, Value: condition}
	}
	p.Consume()
	p.DiscardLineBreaks()

	cases := []MatchCase{}
	stopAt := []TokenKind{RightBrace, EOF}
	for !slices.Contains(stopAt, p.Peek().Kind()) {
		cases = append(cases, parseMatchCase(p))
	}

	next := p.Peek()
	end := next.Loc().End
	if next.Kind() == RightBrace {
		p.Consume()
	} else {
		p.report("'}' expected", next.Loc())
	}
	return MatchExpression{keyword, condition, cases, end}
}

func parseMatchCase(p *Parser) MatchCase {
	var pattern Node
	if p.Peek().Kind() == CaseKeyword {
		p.Consume()
		pattern = ParseExpression(p)
		if p.Peek().Kind() == Colon {
			p.Consume()
		} else {
			p.report("':' expected", p.Peek().Loc())
		}
	}

	stopAt := []TokenKind{EOF, RightBrace, CaseKeyword}
	statements := []Node{}
	for !slices.Contains(stopAt, p.Peek().Kind()) {
		statements = append(statements, p.parseStatement())
		p.DiscardLineBreaks()
	}

	return MatchCase{
		Pattern:    pattern,
		Statements: statements,
	}
}
