package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type BinaryExpression struct {
	Left     Expression
	Right    Expression
	Operator tokenizer.Token
}

func (expr BinaryExpression) Type() ExpressionType {
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
		return expr.Left.Type()
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

func (c *Checker) checkBinaryExpression(expr parser.BinaryExpression) BinaryExpression {
	var left Expression
	if expr.Left != nil {
		left = c.checkExpression(expr.Left)
	}
	var right Expression
	if expr.Right != nil {
		right = c.checkExpression(expr.Right)
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
		c.checkArithmetic(left, right)
	case tokenizer.CONCAT:
		c.checkConcat(left, right)
	case
		tokenizer.LAND,
		tokenizer.LOR:
		c.checkLogical(left, right)
	case
		tokenizer.EQ,
		tokenizer.NEQ:
		c.checkEq(left, right)
	}
	return BinaryExpression{left, right, expr.Operator}
}

func (c *Checker) checkLogical(left Expression, right Expression) {
	if left != nil && !(Primitive{BOOLEAN}).Extends(left.Type()) {
		c.report("The left-hand side of a logical operation must be a boolean", left.Loc())
	}
	if right != nil && !(Primitive{BOOLEAN}).Extends(right.Type()) {
		c.report("The right-hand side of a logical operation must be a boolean", right.Loc())
	}
}
func (c *Checker) checkEq(left Expression, right Expression) {
	if left == nil || right == nil {
		return
	}
	leftType := left.Type()
	rightType := right.Type()
	if !leftType.Extends(rightType) && !rightType.Extends(leftType) {
		c.report("Types don't match", tokenizer.Loc{Start: left.Loc().Start, End: right.Loc().End})
	}
}
func (c *Checker) checkConcat(left Expression, right Expression) {
	var leftType ExpressionType
	if left != nil {
		leftType = left.Type()
	}
	var rightType ExpressionType
	if right != nil {
		rightType = right.Type()
	}
	if leftType != nil && !(Primitive{STRING}).Extends(leftType) && !(List{Primitive{UNKNOWN}}).Extends(leftType) {
		c.report("The left-hand side of concatenation must be a string or a list", left.Loc())
	}
	if rightType != nil && !(Primitive{STRING}).Extends(rightType) && !(List{Primitive{UNKNOWN}}).Extends(rightType) {
		c.report("The right-hand side of concatenation must be a string or a list", right.Loc())
	}
}
func (c *Checker) checkArithmetic(left Expression, right Expression) {
	if left != nil && !(Primitive{NUMBER}).Extends(left.Type()) {
		c.report("The left-hand side of an arithmetic operation must be a number", left.Loc())
	}
	if right != nil && !(Primitive{NUMBER}).Extends(right.Type()) {
		c.report("The right-hand side of an arithmetic operation must be a number", right.Loc())
	}
}
