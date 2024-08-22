package parser

import (
	"slices"

	"github.com/bmelicque/test-parser/tokenizer"
)

type BinaryExpression struct {
	Left     Node
	Right    Node
	Operator tokenizer.Token
}

func (expr BinaryExpression) Loc() tokenizer.Loc {
	loc := tokenizer.Loc{}
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
func parseBinary(p *Parser, operators []tokenizer.TokenKind, fallback func(p *Parser) Node) Node {
	expression := fallback(p)
	next := p.tokenizer.Peek()
	for slices.Contains(operators, next.Kind()) {
		operator := p.tokenizer.Consume()
		right := fallback(p)
		expression = BinaryExpression{expression, right, operator}
		next = p.tokenizer.Peek()
	}
	return expression
}
func parseLogicalOr(p *Parser) Node {
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.LOR}, parseLogicalAnd)
}
func parseLogicalAnd(p *Parser) Node {
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.LAND}, parseEquality)
}
func parseEquality(p *Parser) Node {
	var fallback func(*Parser) Node
	if p.allowAngleBrackets {
		fallback = parseComparison
	} else {
		fallback = parseAddition
	}
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.EQ, tokenizer.NEQ}, fallback)
}
func parseComparison(p *Parser) Node {
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.LESS, tokenizer.LEQ, tokenizer.GEQ, tokenizer.GREATER}, parseAddition)
}
func parseAddition(p *Parser) Node {
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.ADD, tokenizer.CONCAT, tokenizer.SUB}, parseMultiplication)
}
func parseMultiplication(p *Parser) Node {
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.MUL, tokenizer.DIV, tokenizer.MOD}, parseExponentiation)
}
func parseExponentiation(p *Parser) Node {
	expression := p.parseAccessExpression()
	next := p.tokenizer.Peek()
	for next.Kind() == tokenizer.POW {
		operator := p.tokenizer.Consume()
		right := parseExponentiation(p)
		expression = BinaryExpression{expression, right, operator}
		next = p.tokenizer.Peek()
	}
	return expression
}
