package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitBinaryExpression(expr *parser.BinaryExpression) {
	precedence := Precedence(expr)
	if expr.Left != nil {
		left := Precedence(expr.Left)
		if left < precedence {
			e.write("(")
		}
		e.emit(expr.Left)
		if left < precedence {
			e.write(")")
		}
	}

	e.write(" ")
	e.write(expr.Operator.Text())
	e.write(" ")

	if expr.Right != nil {
		right := Precedence(expr.Right)
		if right < precedence {
			e.write("(")
		}
		e.emit(expr.Right)
		if right < precedence {
			e.write(")")
		}
	}
}
