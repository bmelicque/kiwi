package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitBinaryExpression(expr *parser.BinaryExpression) {
	if expr.Operator.Kind() == parser.Equal {
		if e.emitComparison(expr) {
			return
		}
	}

	precedence := Precedence(expr)
	if expr.Left != nil {
		left := Precedence(expr.Left)
		if left < precedence {
			e.write("(")
		}
		e.emitExpression(expr.Left)
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
		e.emitExpression(expr.Right)
		if right < precedence {
			e.write(")")
		}
	}
}

// returns true if wrote something
func (e *Emitter) emitComparison(expr *parser.BinaryExpression) bool {
	switch expr.Left.Type().(type) {
	case parser.List, parser.Ref, parser.Trait, parser.TypeAlias:
		e.write("__.equals(")
		e.emitExpression(expr.Left)
		e.write(", ")
		e.emitExpression(expr.Right)
		e.write(")")
		return true
	}
	return false
}
