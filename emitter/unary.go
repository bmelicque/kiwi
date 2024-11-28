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
		if _, ok := u.Operand.Type().(parser.List); ok {
			e.emitSlice(u.Operand)
		} else {
			e.emitReference(u.Operand)
		}
	case parser.Mul:
		e.emit(u.Operand)
		e.write("()")
	}
}

func (e *Emitter) emitReference(expr parser.Expression) {
	e.write("function (_) { return arguments.length ? void (")
	e.emit(expr)
	e.write(" = _) : ")
	e.emit(expr)
	e.write(" }")
}

func (e *Emitter) emitSlice(expr parser.Expression) {
	var r *parser.RangeExpression
	if computed, ok := expr.(*parser.ComputedAccessExpression); ok {
		expr = computed.Expr
		r = computed.Property.Expr.(*parser.RangeExpression)
	}
	e.addFlag(SliceFlag)
	e.write("__slice(() => ")
	e.emit(expr)
	if r != nil {
		e.write(", ")
		if r.Left != nil {
			e.emit(r.Left)
		} else {
			e.write("0")
		}
		if r.Right != nil {
			e.write(", ")
			e.emit(r.Right)
		}
	}
	e.write(")")
}
