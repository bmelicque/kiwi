package parser

import (
	"fmt"
	"slices"
)

type MatchCase struct {
	Pattern    Expression
	Statements []Node
}

func (m MatchCase) Type() ExpressionType {
	if len(m.Statements) == 0 {
		return Nil{}
	}
	expr, ok := m.Statements[len(m.Statements)-1].(Expression)
	if !ok {
		return Nil{}
	}
	t, _ := expr.Type().build(nil, nil)
	return t
}

func (m MatchCase) IsCatchall() bool {
	identifier, ok := m.Pattern.(*Identifier)
	return ok && identifier.Text() == "_"
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
		if len(m.Cases[i].Statements) > 0 {
			children = append(children, m.Cases[i].Statements...)
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
		return Nil{}
	}
	return m.Cases[0].Type()
}

var matchableType = []ExpressionTypeKind{SUM, TRAIT}

// FIXME: limit possible types for match
func (m *MatchExpression) typeCheck(p *Parser) {
	t := m.Value.Type()
	if t == nil {
		return
	}
	if !slices.Contains(matchableType, t.Kind()) {
		p.report("Cannot match this type", m.Value.Loc())
	}
	for i := range m.Cases {
		p.pushScope(NewScope(BlockScope))
		p.typeCheckPattern(m.Cases[i].Pattern, t)
		for j := range m.Cases[i].Statements {
			m.Cases[i].Statements[j].typeCheck(p)
		}
		p.dropScope()
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
	if p.Peek().Kind() != LeftBrace && !recover(p, LeftBrace) {
		return &MatchExpression{Keyword: keyword, Value: condition}
	}
	p.Consume()
	p.DiscardLineBreaks()

	cases := []MatchCase{}
	stopAt := []TokenKind{RightBrace, EOF}
	for !slices.Contains(stopAt, p.Peek().Kind()) {
		cases = append(cases, parseMatchCase(p))
	}
	validateCaseList(p, cases)

	next := p.Peek()
	end := next.Loc().End
	if next.Kind() == RightBrace {
		p.Consume()
	} else {
		p.report("'}' expected", next.Loc())
	}
	expr := MatchExpression{keyword, condition, cases, end}
	if len(cases) < 2 {
		p.report("At least 2 cases expected", expr.Loc())
	}
	return &expr
}

func parseMatchCase(p *Parser) MatchCase {
	pattern := parseCaseStatement(p)
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

func parseCaseStatement(p *Parser) Expression {
	p.Consume()
	pattern := p.parseExpression()
	if p.Peek().Kind() == Colon || recover(p, Colon) {
		p.Consume()
	}
	if p.Peek().Kind() != EOL {
		recover(p, EOL)
	}
	p.DiscardLineBreaks()
	return pattern
}

func validateCaseList(p *Parser, cases []MatchCase) {
	if len(cases) == 0 {
		return
	}
	if cases[0].Pattern == nil {
		p.report("'case' keyword expected", cases[0].Statements[0].Loc())
	}
	reportUnreachableCases(p, cases)
	reportDuplicatedCases(p, cases)
}

// return true if found a catch-all case
func reportUnreachableCases(p *Parser, cases []MatchCase) {
	var foundCatchall, foundUnreachable bool
	var catchall Loc
	for _, ca := range cases {
		if foundCatchall {
			foundUnreachable = true
		}
		if ca.IsCatchall() {
			foundCatchall = true
			catchall = ca.Pattern.Loc()
		}
	}
	if foundUnreachable {
		p.report("Catch-all case should be last", catchall)
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
		msg := fmt.Sprintf("Duplicated case '%v'", name)
		for _, loc := range locs {
			p.report(msg, loc)
		}
	}
}

// length of cases should be at least 1
func reportMissingCases(p *Parser, cases []MatchCase, matched ExpressionType) {
	last := cases[len(cases)-1]
	loc := Loc{
		cases[0].Pattern.Loc().Start,
		last.Statements[len(last.Statements)-1].Loc().End,
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
		p.report("Non exhaustive match, consider adding catch-all", loc)
	}
	for name := range sum.Members {
		delete(names, name)
	}
	for name := range names {
		p.report(fmt.Sprintf("Missing constructor '%v'", name), loc)
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
