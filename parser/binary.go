package parser

import (
	"slices"
)

type BinaryExpression struct {
	Left     Node
	Right    Node
	Operator Token
}

func (expr BinaryExpression) Loc() Loc {
	loc := Loc{}
	if expr.Left != nil {
		loc.Start = expr.Left.Loc().Start
	} else {
		loc.Start = expr.Operator.Loc().Start
	}

	if expr.Right != nil {
		loc.End = expr.Right.Loc().End
	} else {
		loc.End = expr.Operator.Loc().End
	}
	return loc
}

/******************************
 *  PARSING HELPER FUNCTIONS  *
 ******************************/
func (BinaryExpression) Parse(p *Parser) Node {
	return parseLogicalOr(p)
}
func parseBinary(p *Parser, operators []TokenKind, fallback func(p *Parser) Node) Node {
	expression := fallback(p)
	next := p.Peek()
	for slices.Contains(operators, next.Kind()) {
		operator := p.Consume()
		right := fallback(p)
		expression = BinaryExpression{expression, right, operator}
		next = p.Peek()
	}
	return expression
}
func parseLogicalOr(p *Parser) Node {
	return parseBinary(p, []TokenKind{LOR}, parseLogicalAnd)
}
func parseLogicalAnd(p *Parser) Node {
	return parseBinary(p, []TokenKind{LAND}, parseEquality)
}
func parseEquality(p *Parser) Node {
	return parseBinary(p, []TokenKind{EQ, NEQ}, parseComparison)
}
func parseComparison(p *Parser) Node {
	return parseBinary(p, []TokenKind{LESS, LEQ, GEQ, GREATER}, parseAddition)
}
func parseAddition(p *Parser) Node {
	return parseBinary(p, []TokenKind{ADD, CONCAT, SUB}, parseMultiplication)
}
func parseMultiplication(p *Parser) Node {
	return parseBinary(p, []TokenKind{MUL, DIV, MOD}, parseExponentiation)
}
func parseExponentiation(p *Parser) Node {
	expression := p.parseAccessExpression()
	next := p.Peek()
	for next.Kind() == POW {
		operator := p.Consume()
		right := parseExponentiation(p)
		expression = BinaryExpression{expression, right, operator}
		next = p.Peek()
	}
	return expression
}
