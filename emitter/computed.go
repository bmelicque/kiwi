package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitComputedAccessExpression(expr *parser.ComputedAccessExpression) {
	switch left := expr.Expr.Type().(type) {
	case parser.TypeAlias:
		if left.Name == "Map" {
			emitGetElement(e, expr)
		} else {
			e.emitExpression(expr.Expr)
		}
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
