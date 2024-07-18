package parser

import (
	"slices"

	"github.com/bmelicque/test-parser/tokenizer"
)

type BinaryExpression struct {
	Left     Expression
	Right    Expression
	Operator tokenizer.Token
	loc      tokenizer.Loc
}

func (expr BinaryExpression) Type(ctx *Scope) ExpressionType {
	switch expr.Operator.Kind() {
	case
		tokenizer.ADD,
		tokenizer.SUB,
		tokenizer.MUL,
		tokenizer.POW,
		tokenizer.DIV,
		tokenizer.MOD:
		return Primitive{NUMBER}
	case tokenizer.CONCAT:
		return expr.Left.Type(ctx)
	case
		tokenizer.LAND,
		tokenizer.LOR,
		tokenizer.LESS,
		tokenizer.GREATER,
		tokenizer.LEQ,
		tokenizer.GEQ,
		tokenizer.EQ,
		tokenizer.NEQ:
		return Primitive{BOOLEAN}
	}
	return Primitive{UNKNOWN}
}

func (expr BinaryExpression) Loc() tokenizer.Loc { return expr.loc }

/*******************
 *  TYPE CHECKING  *
 *******************/
func (expr BinaryExpression) Check(c *Checker) {
	if expr.Left != nil {
		expr.Left.Check(c)
	}
	if expr.Right != nil {
		expr.Right.Check(c)
	}
	switch expr.Operator.Kind() {
	case
		tokenizer.ADD,
		tokenizer.SUB,
		tokenizer.MUL,
		tokenizer.POW,
		tokenizer.DIV,
		tokenizer.MOD,
		tokenizer.LESS,
		tokenizer.GREATER,
		tokenizer.LEQ,
		tokenizer.GEQ:
		expr.checkArithmetic(c)
	case tokenizer.CONCAT:
		expr.checkConcat(c)
	case
		tokenizer.LAND,
		tokenizer.LOR:
		expr.checkEq(c)
	case
		tokenizer.EQ,
		tokenizer.NEQ:
		// TODO: check that types overlap
	}
}

func (expr BinaryExpression) checkLogical(c *Checker) {
	if expr.Left != nil && !(Primitive{BOOLEAN}).Extends(expr.Left.Type(c.scope)) {
		c.report("The left-hand side of a logical operation must be a boolean", expr.Left.Loc())
	}
	if expr.Right != nil && !(Primitive{BOOLEAN}).Extends(expr.Right.Type(c.scope)) {
		c.report("The right-hand side of a logical operation must be a boolean", expr.Right.Loc())
	}
}
func (expr BinaryExpression) checkEq(c *Checker) {
	if expr.Left == nil || expr.Right == nil {
		return
	}
	left := expr.Left.Type(c.scope)
	right := expr.Right.Type(c.scope)
	if !left.Extends(right) && !right.Extends(left) {
		c.report("Types don't match", expr.Loc())
	}
}
func (expr BinaryExpression) checkConcat(c *Checker) {
	var left ExpressionType
	if expr.Left != nil {
		left = expr.Left.Type(c.scope)
	}
	var right ExpressionType
	if expr.Right != nil {
		right = expr.Right.Type(c.scope)
	}
	if left != nil && !(Primitive{STRING}).Extends(left) && !(List{Primitive{UNKNOWN}}).Extends(left) {
		c.report("The left-hand side of concatenation must be a string or a list", expr.Left.Loc())
	}
	if right != nil && !(Primitive{STRING}).Extends(right) && !(List{Primitive{UNKNOWN}}).Extends(right) {
		c.report("The right-hand side of concatenation must be a string or a list", expr.Right.Loc())
	}
}
func (expr BinaryExpression) checkArithmetic(c *Checker) {
	if expr.Left != nil && !(Primitive{NUMBER}).Extends(expr.Left.Type(c.scope)) {
		c.report("The left-hand side of an arithmetic operation must be a number", expr.Left.Loc())
	}
	if expr.Right != nil && !(Primitive{NUMBER}).Extends(expr.Right.Type(c.scope)) {
		c.report("The right-hand side of an arithmetic operation must be a number", expr.Right.Loc())
	}
}

/******************************
 *  PARSING HELPER FUNCTIONS  *
 ******************************/
func (BinaryExpression) Parse(p *Parser) Expression {
	return parseLogicalOr(p)
}
func parseBinary(p *Parser, operators []tokenizer.TokenKind, fallback func(p *Parser) Expression) Expression {
	expression := fallback(p)
	next := p.tokenizer.Peek()
	for slices.Contains(operators, next.Kind()) {
		operator := p.tokenizer.Consume()
		right := fallback(p)
		expression = BinaryExpression{expression, right, operator, tokenizer.Loc{}}.setLoc()
		next = p.tokenizer.Peek()
	}
	return expression
}
func parseLogicalOr(p *Parser) Expression {
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.LOR}, parseLogicalAnd)
}
func parseLogicalAnd(p *Parser) Expression {
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.LAND}, parseEquality)
}
func parseEquality(p *Parser) Expression {
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.EQ, tokenizer.NEQ}, parseComparison)
}
func parseComparison(p *Parser) Expression {
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.LESS, tokenizer.LEQ, tokenizer.GEQ, tokenizer.GREATER}, parseAddition)
}
func parseAddition(p *Parser) Expression {
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.ADD, tokenizer.CONCAT, tokenizer.SUB}, parseMultiplication)
}
func parseMultiplication(p *Parser) Expression {
	return parseBinary(p, []tokenizer.TokenKind{tokenizer.MUL, tokenizer.DIV, tokenizer.MOD}, parseExponentiation)
}
func parseExponentiation(p *Parser) Expression {
	expression := fallback(p)
	next := p.tokenizer.Peek()
	for next.Kind() == tokenizer.POW {
		operator := p.tokenizer.Consume()
		right := parseExponentiation(p)
		expression = BinaryExpression{expression, right, operator, tokenizer.Loc{}}.setLoc()
		next = p.tokenizer.Peek()
	}
	return expression
}
func fallback(p *Parser) Expression {
	if p.tokenizer.Peek().Kind() == tokenizer.LPAREN {
		return ParseFunctionExpression(p)
	}
	if p.tokenizer.Peek().Kind() == tokenizer.LBRACKET {
		return ListExpression{}.Parse(p)
	}
	return TokenExpression{}.Parse(p)
}
func (expr BinaryExpression) setLoc() BinaryExpression {
	if expr.Left != nil {
		expr.loc.Start = expr.Left.Loc().Start
	} else {
		expr.loc.Start = expr.Operator.Loc().Start
	}

	if expr.Right != nil {
		expr.loc.End = expr.Right.Loc().End
	} else {
		expr.loc.End = expr.Operator.Loc().End
	}
	return expr
}
