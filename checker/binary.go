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

/*******************
 *  TYPE CHECKING  *
 *******************/
func (c *Checker) Check(expr parser.BinaryExpression) BinaryExpression {
	var left Expression
	if expr.Left != nil {
		left = c.CheckExpression(expr.Left)
	}
	var right Expression
	if expr.Right != nil {
		right = c.CheckExpression(expr.Right)
	}

	binary := BinaryExpression{left, right, expr.Operator}
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
		c.checkArithmetic(binary)
	case tokenizer.CONCAT:
		c.checkConcat(binary)
	case
		tokenizer.LAND,
		tokenizer.LOR:
		c.checkEq(binary)
	case
		tokenizer.EQ,
		tokenizer.NEQ:
		// TODO: check that types overlap
	}
	return binary
}

func (c *Checker) checkLogical(expr BinaryExpression) {
	if expr.Left != nil && !(Primitive{BOOLEAN}).Extends(expr.Left.Type()) {
		c.report("The left-hand side of a logical operation must be a boolean", expr.Left.Loc())
	}
	if expr.Right != nil && !(Primitive{BOOLEAN}).Extends(expr.Right.Type()) {
		c.report("The right-hand side of a logical operation must be a boolean", expr.Right.Loc())
	}
}
func (c *Checker) checkEq(expr BinaryExpression) {
	if expr.Left == nil || expr.Right == nil {
		return
	}
	left := expr.Left.Type()
	right := expr.Right.Type()
	if !left.Extends(right) && !right.Extends(left) {
		c.report("Types don't match", expr.Loc())
	}
}
func (c *Checker) checkConcat(expr BinaryExpression) {
	var left ExpressionType
	if expr.Left != nil {
		left = expr.Left.Type()
	}
	var right ExpressionType
	if expr.Right != nil {
		right = expr.Right.Type()
	}
	if left != nil && !(Primitive{STRING}).Extends(left) && !(List{Primitive{UNKNOWN}}).Extends(left) {
		c.report("The left-hand side of concatenation must be a string or a list", expr.Left.Loc())
	}
	if right != nil && !(Primitive{STRING}).Extends(right) && !(List{Primitive{UNKNOWN}}).Extends(right) {
		c.report("The right-hand side of concatenation must be a string or a list", expr.Right.Loc())
	}
}
func (c *Checker) checkArithmetic(expr BinaryExpression) {
	if expr.Left != nil && !(Primitive{NUMBER}).Extends(expr.Left.Type()) {
		c.report("The left-hand side of an arithmetic operation must be a number", expr.Left.Loc())
	}
	if expr.Right != nil && !(Primitive{NUMBER}).Extends(expr.Right.Type()) {
		c.report("The right-hand side of an arithmetic operation must be a number", expr.Right.Loc())
	}
}
