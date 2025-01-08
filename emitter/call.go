package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitCallExpression(expr *parser.CallExpression, await bool) {
	args := expr.Args.Expr.(*parser.TupleExpression).Elements
	if isNodeMethod(expr.Callee) && len(args) == 0 {
		emitNodeGetterCall(e, expr)
		return
	}
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

func isNodeMethod(expr parser.Expression) bool {
	p, ok := expr.(*parser.PropertyAccessExpression)
	return ok && implementsNode(p.Expr.Type())
}

func emitNodeGetterCall(e *Emitter, expr *parser.CallExpression) {
	e.write("__.callNodeGetter(")
	p := expr.Callee.(*parser.PropertyAccessExpression)
	e.emitExpression(p.Expr)
	e.write(", \"")
	e.emitExpression(p.Property)
	e.write("\")")
}
