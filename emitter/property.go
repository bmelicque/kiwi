package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitPropertyAccessExpression(p *parser.PropertyAccessExpression, isCalled bool) {
	e.emitExpression(p.Expr)
	if _, isRef := p.Expr.Type().(parser.Ref); isRef {
		e.write("(1)")
	}
	if _, ok := p.Expr.Type().(parser.Tuple); ok {
		e.write("[")
		e.emitExpression(p.Property)
		e.write("]")
	} else {
		e.write(".")
		e.emitExpression(p.Property)
	}
	// a method which is not called has to be bound to handle correct behavior of "this"
	if _, ok := p.Type().(parser.Function); ok && !isCalled {
		// FIXME: if expr is expensive, should be computed only once
		e.write(".bind(")
		e.emitExpression(p.Expr)
		e.write(")")
	}
}
