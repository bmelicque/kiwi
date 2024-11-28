package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitPropertyAccessExpression(p *parser.PropertyAccessExpression) {
	e.emit(p.Expr)
	if _, ok := p.Expr.Type().(parser.Tuple); ok {
		e.write("[")
		e.emit(p.Property)
		e.write("]")
	} else {
		e.write(".")
		e.emit(p.Property)
	}
}
