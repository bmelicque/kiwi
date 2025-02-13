package parser

import (
	"slices"
)

type MatchCase struct {
	Pattern    Expression
	Colon      Token
	Consequent Expression
}

func (m MatchCase) Type() ExpressionType {
	if m.Consequent == nil {
		return Invalid{}
	}

	return m.Consequent.Type()
}

func (m *MatchCase) typeCheck(p *Parser, pattern ExpressionType) {
	p.pushScope(NewScope(BlockScope))
	p.typeCheckPattern(m.Pattern, pattern)
	if m.Consequent != nil {
		m.Consequent.typeCheck(p)
	}
	p.dropScope()
}

func (m MatchCase) IsCatchall() bool {
	identifier, ok := m.Pattern.(*Identifier)
	return ok && identifier.Text() == "_"
}

func (m MatchCase) Loc() Loc {
	var start, end Position
	if m.Pattern != nil {
		start = m.Pattern.Loc().Start
	} else {
		start = m.Colon.Loc().Start
	}
	if m.Consequent != nil {
		end = m.Consequent.Loc().End
	} else {
		end = m.Colon.Loc().End
	}
	return Loc{start, end}
}

func parseMatchCase(p *Parser) MatchCase {
	pattern := p.parseTaggedExpression()
	switch pattern := pattern.(type) {
	case *Entry:
		if pattern.Value == nil {
			p.error(&Literal{p.Peek()}, ExpressionExpected)
		}
		return MatchCase{
			Pattern:    pattern.Key,
			Colon:      pattern.Colon,
			Consequent: pattern.Value,
		}
	default:
		// Note: this default includes *Param as a valid pattern
		if p.Peek().Kind() != Colon && !recoverBadTokens(p, Colon) {
			return MatchCase{Pattern: pattern}
		}
		colon := p.Consume()
		consequent := p.parseExpression()
		if consequent == nil {
			p.error(&Literal{p.Peek()}, ExpressionExpected)
		}
		return MatchCase{
			Pattern:    pattern,
			Colon:      colon,
			Consequent: consequent,
		}
	}
}

type MatchExpression struct {
	Keyword Token
	Value   Expression
	Cases   []MatchCase
	end     Position
}

func (m *MatchExpression) getChildren() []Node {
	children := []Node{}
	if m.Value != nil {
		children = append(children, m.Value)
	}
	for i := range m.Cases {
		if m.Cases[i].Pattern != nil {
			children = append(children, m.Cases[i].Pattern)
		}
		if m.Cases[i].Consequent != nil {
			children = append(children, m.Cases[i].Consequent)
		}
	}
	return children
}

func (m *MatchExpression) Loc() Loc {
	loc := m.Keyword.Loc()
	if m.end != (Position{}) {
		loc.End = m.end
	}
	return loc
}
func (m *MatchExpression) Type() ExpressionType {
	if len(m.Cases) == 0 {
		return Void{}
	}
	return m.Cases[0].Type()
}

func (m *MatchExpression) typeCheck(p *Parser) {
	m.Value.typeCheck(p)
	t := m.Value.Type()
	if t == nil {
		return
	}
	if alias, ok := t.(TypeAlias); ok {
		t = alias.Ref
	}
	switch t.(type) {
	case Sum, Trait:
	default:
		p.error(m.Value, Unmatchable, t)
	}
	for i := range m.Cases {
		m.Cases[i].typeCheck(p, t)
	}
	reportMissingCases(p, m.Cases, t)
}

// TODO: validate type
func (p *Parser) parseMatchExpression() Expression {
	keyword := p.Consume()
	outer := p.allowBraceParsing
	p.allowBraceParsing = false
	condition := p.parseExpression()
	p.allowBraceParsing = outer
	if p.Peek().Kind() != LeftBrace && !recoverBadTokens(p, LeftBrace) {
		return &MatchExpression{Keyword: keyword, Value: condition}
	}
	p.Consume()
	p.DiscardLineBreaks()

	cases := []MatchCase{}
	stopAt := []TokenKind{RightBrace, EOF}
	for !slices.Contains(stopAt, p.Peek().Kind()) {
		cases = append(cases, parseMatchCase(p))
		if p.Peek().Kind() != EOL {
			recoverBadTokens(p, EOL)
		}
		p.DiscardLineBreaks()
	}
	validateCaseList(p, cases)

	next := p.Peek()
	end := next.Loc().End
	if next.Kind() == RightBrace {
		p.Consume()
	} else {
		p.error(&Literal{next}, RightBraceExpected)
	}
	expr := MatchExpression{keyword, condition, cases, end}
	if len(cases) < 2 {
		p.error(&expr, MissingElements, "at least 2", len(cases))
	}
	return &expr
}

func validateCaseList(p *Parser, cases []MatchCase) {
	if len(cases) == 0 {
		return
	}
	reportUnreachableCases(p, cases)
	reportDuplicatedCases(p, cases)
}

// return true if found a catch-all case
func reportUnreachableCases(p *Parser, cases []MatchCase) {
	var foundCatchall, foundUnreachable bool
	var catchall Expression
	for _, ca := range cases {
		if foundCatchall {
			foundUnreachable = true
		}
		if ca.IsCatchall() {
			foundCatchall = true
			catchall = ca.Pattern
		}
	}
	if foundUnreachable {
		p.error(catchall, CatchallNotLast)
	}
}

func reportDuplicatedCases(p *Parser, cases []MatchCase) {
	names := map[string][]Loc{}
	for _, c := range cases {
		identifier := getCaseIdentifier(c)
		if identifier != nil {
			name := identifier.Text()
			names[name] = append(names[name], identifier.Loc())
		}
	}
	for name, locs := range names {
		if len(locs) == 1 {
			continue
		}
		for _, loc := range locs {
			p.error(&Block{loc: loc}, DuplicateIdentifier, name)
		}
	}
}

// length of cases should be at least 1
func reportMissingCases(p *Parser, cases []MatchCase, matched ExpressionType) {
	last := cases[len(cases)-1]
	loc := Loc{
		cases[0].Pattern.Loc().Start,
		last.Loc().End,
	}
	names := map[string]bool{}
	for _, c := range cases {
		identifier := getCaseIdentifier(c)
		if identifier == nil {
			continue
		}
		if identifier.Text() == "_" {
			// No missing case if catch-all is found
			return
		}
		names[identifier.Text()] = true
	}
	sum, ok := matched.(Sum)
	if !ok {
		p.error(&Block{loc: loc}, NotExhaustive)
	}
	for name := range sum.Members {
		delete(names, name)
	}
	b := &Block{loc: loc}
	for name := range names {
		p.error(b, MissingConstructor, name)
	}
}

func getCaseIdentifier(c MatchCase) *Identifier {
	switch pattern := c.Pattern.(type) {
	case *Identifier:
		return pattern
	case *CallExpression:
		identifier, _ := pattern.Callee.(*Identifier)
		return identifier
	default:
		return nil
	}
}
