package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitCallExpression(expr *parser.CallExpression, await bool) {
	if expr.Callee.Type().(parser.Function).Async && await {
		e.write("await ")
	}
	e.emitExpression(expr.Callee)
	e.write("(")
	defer e.write(")")

	args := expr.Args.Expr.(*parser.TupleExpression).Elements
	max := len(args) - 1
	if max == -1 {
		return
	}
	for i := range args[:max] {
		e.emitExpression(args[i])
		e.write(", ")
	}
	e.emitExpression(args[max])
}
