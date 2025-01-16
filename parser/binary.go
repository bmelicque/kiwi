package parser

import (
	"fmt"
	"slices"
)

type BinaryExpression struct {
	Left     Expression
	Right    Expression
	Operator Token
}

func (b *BinaryExpression) getChildren() []Node {
	children := []Node{}
	if b.Left != nil {
		children = append(children, b.Left)
	}
	if b.Right != nil {
		children = append(children, b.Right)
	}
	return children
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
		return Number{}
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
		return Boolean{}
	case Bang:
		left := expr.Left.Type()
		if t, ok := left.(Type); ok {
			left = t.Value
		} else {
			left = Unknown{}
		}

		right := expr.Right.Type()
		if t, ok := right.(Type); ok {
			right = t.Value
		} else {
			right = Unknown{}
		}
		return Type{makeResultType(right, left)}
	case InKeyword:
		return Void{}
	default:
		panic(fmt.Sprintf("operator '%v' not implemented", expr.Operator.Kind()))
	}
}

/******************************
 *  PARSING HELPER FUNCTIONS  *
 ******************************/
func (p *Parser) parseBinaryExpression() Expression {
	return parseBinaryErrorType(p)
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
func parseBinaryErrorType(p *Parser) Expression {
	return parseBinary(p, []TokenKind{Bang}, parseLogicalOr)
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
	expression := p.parseCatchExpression()
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
	case Bang:
		typeCheckBinaryErrorType(p, b.Left, b.Right)
	default:
		panic(fmt.Sprintf("operator '%v' not implemented", b.Operator.Kind()))
	}
}

func (p *Parser) typeCheckLogicalExpression(left Expression, right Expression) {
	if left != nil && !(Boolean{}).Extends(left.Type()) {
		p.error(left, BooleanExpected, left.Type())
	}
	if right != nil && !(Boolean{}).Extends(right.Type()) {
		p.error(right, BooleanExpected, right.Type())
	}
}

func (p *Parser) typeCheckComparisonExpression(left Expression, right Expression) {
	if left == nil || right == nil {
		return
	}
	leftType := left.Type()
	rightType := right.Type()
	if !Match(leftType, rightType) {
		dummy := &BinaryExpression{Left: left, Right: right}
		p.error(dummy, MismatchedTypes, leftType, rightType)
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
	if leftType != nil && !(String{}).Extends(leftType) && !(List{Unknown{}}).Extends(leftType) {
		p.error(left, ConcatenableExpected, left.Type())
	}
	if rightType != nil && !(String{}).Extends(rightType) && !(List{Unknown{}}).Extends(rightType) {
		p.error(right, ConcatenableExpected, right.Type())
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
		dummy := &BinaryExpression{Left: left, Right: right}
		p.error(dummy, MismatchedTypes, leftType, rightType)
	}
}
func (p *Parser) typeCheckArithmeticExpression(left Expression, right Expression) {
	if left != nil && !(Number{}).Extends(left.Type()) {
		p.error(left, NumberExpected, left.Type())
	}
	if right != nil && !(Number{}).Extends(right.Type()) {
		p.error(right, NumberExpected, right.Type())
	}
}
func typeCheckBinaryErrorType(p *Parser, left Expression, right Expression) {
	if _, ok := left.Type().(Type); !ok {
		p.error(left, TypeExpected)
	}
	if _, ok := right.Type().(Type); !ok {
		p.error(right, TypeExpected)
	}
}
