package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitUnaryExpression(u *parser.UnaryExpression) {
	switch u.Operator.Kind() {
	case parser.AsyncKeyword:
		e.emitCallExpression(u.Operand.(*parser.CallExpression), false)
	case parser.AwaitKeyword:
		e.write("await ")
		e.emitExpression(u.Operand)
	case parser.Bang:
		e.write("!")
		e.emitExpression(u.Operand)
	case parser.TryKeyword:
		e.emitExpression(u.Operand)
	case parser.BinaryAnd:
		e.emitReference(u.Operand)
	case parser.Mul:
		e.emitExpression(u.Operand)
		if _, ok := u.Operand.Type().(parser.Ref).To.(parser.List); ok {
			e.write(".clone()")
		} else {
			e.write("(1)")
		}
	}
}

func (e *Emitter) emitReference(expr parser.Expression) {
	switch expr := expr.(type) {
	case *parser.PropertyAccessExpression:
		e.emitObjectReference(expr)
	default:
		e.emitPrimitiveReference(expr)
	}
}

func (e *Emitter) emitPrimitiveReference(expr parser.Expression) {
	e.write("(_,__)=>(_&4?__s:_&2?\"")
	e.emitExpression(expr)
	e.write("\":_?")
	e.emitExpression(expr)
	e.write(":(")
	e.emitExpression(expr)
	e.write("=__))")
}
func (e *Emitter) emitObjectReference(expr *parser.PropertyAccessExpression) {
	e.write("__ptr(")
	e.emitExpression(expr.Expr)
	e.write(",\"")
	e.emitExpression(expr.Property)
	e.write("\")")
}
