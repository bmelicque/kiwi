package parser

import (
	"slices"
)

type BinaryExpression struct {
	Left     Expression
	Right    Expression
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

func (expr BinaryExpression) Type() ExpressionType {
	switch expr.Operator.Kind() {
	case
		Add,
		Sub,
		Mul,
		Pow,
		Div,
		Mod:
		return Primitive{NUMBER}
	case Concat:
		return expr.Left.Type()
	case
		LogicalAnd,
		LogicalOr,
		Less,
		Greater,
		LessEqual,
		GreaterEqual,
		Equal,
		NotEqual:
		return Primitive{BOOLEAN}
	}
	return Primitive{UNKNOWN}
}

/******************************
 *  PARSING HELPER FUNCTIONS  *
 ******************************/
func (BinaryExpression) Parse(p *Parser) Expression {
	return parseLogicalOr(p)
}
func parseBinary(p *Parser, operators []TokenKind, fallback func(p *Parser) Expression) Expression {
	expression := fallback(p)
	next := p.Peek()
	for slices.Contains(operators, next.Kind()) {
		operator := p.Consume()
		right := fallback(p)
		expression = &BinaryExpression{expression, right, operator}
		next = p.Peek()
	}
	return expression
}
func parseLogicalOr(p *Parser) Expression {
	expr := parseBinary(p, []TokenKind{LogicalOr}, parseLogicalAnd)
	binary, ok := expr.(*BinaryExpression)
	if ok {
		p.validateBinaryExpression(*binary)
	}
	return expr
}
func parseLogicalAnd(p *Parser) Expression {
	return parseBinary(p, []TokenKind{LogicalAnd}, parseEquality)
}
func parseEquality(p *Parser) Expression {
	return parseBinary(p, []TokenKind{Equal, NotEqual}, parseComparison)
}
func parseComparison(p *Parser) Expression {
	return parseBinary(p, []TokenKind{Less, LessEqual, GreaterEqual, Greater}, parseAddition)
}
func parseAddition(p *Parser) Expression {
	return parseBinary(p, []TokenKind{Add, Concat, Sub}, parseMultiplication)
}
func parseMultiplication(p *Parser) Expression {
	return parseBinary(p, []TokenKind{Mul, Div, Mod}, parseExponentiation)
}
func parseExponentiation(p *Parser) Expression {
	expression := p.parseAccessExpression()
	next := p.Peek()
	for next.Kind() == Pow {
		operator := p.Consume()
		right := parseExponentiation(p)
		expression = &BinaryExpression{expression, right, operator}
		next = p.Peek()
	}
	return expression
}

func (p *Parser) validateBinaryExpression(expr BinaryExpression) {
	left := expr.Left
	right := expr.Right
	switch expr.Operator.Kind() {
	case
		Add,
		Sub,
		Mul,
		Pow,
		Div,
		Mod,
		Less,
		Greater,
		LessEqual,
		GreaterEqual:
		p.validateArithmeticExpression(left, right)
	case Concat:
		p.validateConcatExpression(left, right)
	case
		LogicalAnd,
		LogicalOr:
		p.validateLogicalExpression(left, right)
	case
		Equal,
		NotEqual:
		p.validateComparisonExpression(left, right)
	}
}

func (p *Parser) validateLogicalExpression(left Expression, right Expression) {
	if left != nil && !(Primitive{BOOLEAN}).Extends(left.Type()) {
		p.report("The left-hand side of a logical operation must be a boolean", left.Loc())
	}
	if right != nil && !(Primitive{BOOLEAN}).Extends(right.Type()) {
		p.report("The right-hand side of a logical operation must be a boolean", right.Loc())
	}
}

func (p *Parser) validateComparisonExpression(left Expression, right Expression) {
	if left == nil || right == nil {
		return
	}
	leftType := left.Type()
	rightType := right.Type()
	if !Match(leftType, rightType) {
		p.report("Types don't match", Loc{Start: left.Loc().Start, End: right.Loc().End})
	}
}
func (p *Parser) validateConcatExpression(left Expression, right Expression) {
	var leftType ExpressionType
	if left != nil {
		leftType = left.Type()
	}
	var rightType ExpressionType
	if right != nil {
		rightType = right.Type()
	}
	if leftType != nil && !(Primitive{STRING}).Extends(leftType) && !(List{Primitive{UNKNOWN}}).Extends(leftType) {
		p.report("The left-hand side of concatenation must be a string or a list", left.Loc())
	}
	if rightType != nil && !(Primitive{STRING}).Extends(rightType) && !(List{Primitive{UNKNOWN}}).Extends(rightType) {
		p.report("The right-hand side of concatenation must be a string or a list", right.Loc())
	}
}
func (p *Parser) validateArithmeticExpression(left Expression, right Expression) {
	if left != nil && !(Primitive{NUMBER}).Extends(left.Type()) {
		p.report("The left-hand side of an arithmetic operation must be a number", left.Loc())
	}
	if right != nil && !(Primitive{NUMBER}).Extends(right.Type()) {
		p.report("The right-hand side of an arithmetic operation must be a number", right.Loc())
	}
}
