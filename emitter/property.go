package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitPropertyAccessExpression(p *parser.PropertyAccessExpression) {
	e.emitExpression(p.Expr)
	if _, ok := p.Expr.Type().(parser.Tuple); ok {
		e.write("[")
		e.emitExpression(p.Property)
		e.write("]")
	} else {
		e.write(".")
		e.emitExpression(p.Property)
	}
}
