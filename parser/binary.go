package parser

import (
	"slices"
)

type BinaryExpression struct {
	Left     Expression
	Right    Expression
	Operator Token
}

func (b *BinaryExpression) Walk(cb func(Node), skip func(Node) bool) {
	if skip(b) {
		return
	}
	cb(b)
	if b.Left != nil {
		b.Left.Walk(cb, skip)
	}
	if b.Right != nil {
		b.Right.Walk(cb, skip)
	}
}

func (expr *BinaryExpression) Loc() Loc {
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

func (expr *BinaryExpression) Type() ExpressionType {
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
func (p *Parser) parseBinaryExpression() Expression {
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
	return parseBinary(p, []TokenKind{LogicalOr}, parseLogicalAnd)
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

func (b *BinaryExpression) typeCheck(p *Parser) {
	b.Left.typeCheck(p)
	b.Right.typeCheck(p)
	switch b.Operator.Kind() {
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
		p.typeCheckArithmeticExpression(b.Left, b.Right)
	case Concat:
		p.typeCheckConcatExpression(b.Left, b.Right)
	case
		LogicalAnd,
		LogicalOr:
		p.typeCheckLogicalExpression(b.Left, b.Right)
	case
		Equal,
		NotEqual:
		p.typeCheckComparisonExpression(b.Left, b.Right)
	}
}

func (p *Parser) typeCheckLogicalExpression(left Expression, right Expression) {
	if left != nil && !(Primitive{BOOLEAN}).Extends(left.Type()) {
		p.report("The left-hand side of a logical operation must be a boolean", left.Loc())
	}
	if right != nil && !(Primitive{BOOLEAN}).Extends(right.Type()) {
		p.report("The right-hand side of a logical operation must be a boolean", right.Loc())
	}
}

func (p *Parser) typeCheckComparisonExpression(left Expression, right Expression) {
	if left == nil || right == nil {
		return
	}
	leftType := left.Type()
	rightType := right.Type()
	if !Match(leftType, rightType) {
		p.report("Types don't match", Loc{Start: left.Loc().Start, End: right.Loc().End})
	}
}
func (p *Parser) typeCheckConcatExpression(left Expression, right Expression) {
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

	rightList, ok := rightType.(List)
	if !ok {
		return
	}
	leftList, ok := leftType.(List)
	if !ok {
		return
	}
	if !leftList.Element.Extends(rightList.Element) {
		p.report("Element type doesn't match lhs", right.Loc())
	}
}
func (p *Parser) typeCheckArithmeticExpression(left Expression, right Expression) {
	if left != nil && !(Primitive{NUMBER}).Extends(left.Type()) {
		p.report("The left-hand side of an arithmetic operation must be a number", left.Loc())
	}
	if right != nil && !(Primitive{NUMBER}).Extends(right.Type()) {
		p.report("The right-hand side of an arithmetic operation must be a number", right.Loc())
	}
}
