package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitListAccess(expr *parser.ComputedAccessExpression) {
	switch prop := expr.Property.Expr.(type) {
	case *parser.RangeExpression:
		e.write(".slice(")
		if prop.Left != nil {
			e.emitExpression(prop.Left)
		} else {
			e.write("0")
		}
		if prop.Right == nil {
			e.write(")")
			return
		}
		e.write(", ")
		e.emitExpression(prop.Right)
		if prop.Operator.Kind() == parser.InclusiveRange {
			e.write("+1")
		}
		e.write(")")
	default:
		e.write("[")
		e.emitExpression(expr.Property.Expr)
		e.write("]")
	}
}
func (e *Emitter) emitComputedAccessExpression(expr *parser.ComputedAccessExpression) {
	switch left := expr.Expr.Type().(type) {
	case parser.TypeAlias:
		if left.Name == "Map" {
			emitGetElement(e, expr)
		} else {
			e.emitExpression(expr.Expr)
		}
	case parser.Ref:
		if _, ok := left.To.(parser.List); !ok {
			panic("unexpected typing (expected &[]any)")
		}
		emitGetElement(e, expr)
	case parser.List:
		e.emitListAccess(expr)
	default:
		e.emitExpression(expr.Expr)
	}
}
func emitGetElement(e *Emitter, c *parser.ComputedAccessExpression) {
	e.emitExpression(c.Expr)
	e.write(".get(")
	e.emitExpression(c.Property.Expr)
	e.write(")")
}
