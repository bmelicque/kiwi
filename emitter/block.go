package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

func emitScope(e *Emitter, scope *parser.Scope) {
	e.write(fmt.Sprintf("__s%v", scope.GetId()))
}

func (e *Emitter) emitBlockStatement(b *parser.Block) {
	e.write("{")
	if len(b.Statements) == 0 {
		e.write("}")
		return
	}
	e.write("\n")
	e.depth++
	if needsSymbol(b) {
		e.indent()
		e.write("const __s = Symbol();\n")
	}
	if b.Scope().HasReferencedVars() {
		e.indent()
		e.write("const ")
		emitScope(e, b.Scope())
	}
	for _, statement := range b.Statements {
		e.indent()
		e.emit(statement)
	}
	e.depth--
	e.indent()
	e.write("}\n")
}

func needsSymbol(b *parser.Block) bool {
	var found bool
	parser.Walk(b, func(n parser.Node, skip func()) {
		if found {
			skip()
			return
		}
		unary, ok := n.(*parser.UnaryExpression)
		if !ok || unary.Operator.Kind() != parser.BinaryAnd {
			return
		}
		if _, ok = unary.Operand.(*parser.Identifier); ok {
			found = true
			skip()
		}
	})
	return found
}

func (e *Emitter) emitBlockExpression(b *parser.Block) {
	if id, ok := e.uninlinables[b]; ok {
		e.write(fmt.Sprintf("_tmp%v", id))
		delete(e.uninlinables, b)
		return
	}

	if len(b.Statements) == 0 {
		e.write("undefined")
		return
	}
	if len(b.Statements) == 1 {
		e.emitExpression(b.Statements[0].(parser.Expression))
		return
	}
	e.write("(\n")
	e.depth += 1
	for _, statement := range b.Statements {
		e.indent()
		e.emitExpression(statement.(parser.Expression))
		e.write(",\n")
	}
	e.depth -= 1
	e.indent()
	e.write(")")
}
