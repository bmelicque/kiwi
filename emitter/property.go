package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitPropertyAccessExpression(p *parser.PropertyAccessExpression, isCalled bool) {
	if _, ok := p.Type().(parser.Function); ok && !isCalled {
		emitNoncalledMethod(e, p)
		return
	}

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

}

// A method which is not called has to be bound to handle correct behavior of "this".
// Method on Node trait has to have extra handling.
func emitNoncalledMethod(e *Emitter, p *parser.PropertyAccessExpression) {
	if isNodeMethod(p) {
		e.write("__.bindNodeGetter(")
	} else {
		e.write("__.bind(")
	}
	e.emitExpression(p.Expr)
	e.write(", \"")
	e.emitExpression(p.Property)
	e.write("\")")
}
