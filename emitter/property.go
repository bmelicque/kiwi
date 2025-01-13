package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitPropertyAccessExpression(p *parser.PropertyAccessExpression, isCalled bool) {
	_, isMethod := p.Type().(parser.Function)
	switch {
	case isMethod && implementsNode(p.Expr.Type()):
		emitNodeMethod(e, p)
	case isMethod && !isCalled:
		emitNoncalledMethod(e, p)
	default:
		emitPropertyAccess(e, p)
	}
}

// Native methods on nodes are tricky to handle (concerning refs for example)
func emitNodeMethod(e *Emitter, p *parser.PropertyAccessExpression) {
	e.write("__.wrapNodeMethod(")
	e.emitExpression(p.Expr)
	e.write(", \"")
	e.emitExpression(p.Property)
	f := p.Type().(parser.Function)
	if _, returnsNode := f.Returned.(parser.Ref); returnsNode {
		e.write("\", 1)")
	} else {
		e.write("\", 0)")
	}
}

// A method which is not called has to be bound to handle correct behavior of "this".
// Method on Node trait has to have extra handling.
func emitNoncalledMethod(e *Emitter, p *parser.PropertyAccessExpression) {
	e.write("__.bind(")
	e.emitExpression(p.Expr)
	e.write(", \"")
	e.emitExpression(p.Property)
	e.write("\")")
}

func emitPropertyAccess(e *Emitter, p *parser.PropertyAccessExpression) {
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
