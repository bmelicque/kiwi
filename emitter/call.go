package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitCallExpression(expr *parser.CallExpression, await bool) {
	args := expr.Args.Expr.(*parser.TupleExpression).Elements
	if expr.Callee.Type().(parser.Function).Async && await {
		e.write("await ")
	}

	if p, ok := expr.Callee.(*parser.PropertyAccessExpression); ok {
		e.emitPropertyAccessExpression(p, true)
	} else {
		e.emitExpression(expr.Callee)
	}

	max := len(args) - 1
	if max == -1 {
		e.write("()")
		return
	}

	e.write("(")
	for i := range args[:max] {
		e.emitExpression(args[i])
		e.write(", ")
	}
	e.emitExpression(args[max])
	e.write(")")
}
