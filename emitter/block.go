package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

func (e *Emitter) emitBlockStatement(b *parser.Block) {
	e.write("{")
	if len(b.Statements) == 0 {
		e.write("}")
		return
	}
	e.write("\n")
	e.depth++
	for _, statement := range b.Statements {
		e.indent()
		e.emit(statement)
		if _, ok := statement.(parser.Expression); ok {
			e.write(";\n")
		}
	}
	e.depth--
	e.indent()
	e.write("}\n")
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
		e.emit(b.Statements[0])
		return
	}
	e.write("(\n")
	e.depth += 1
	for _, statement := range b.Statements {
		e.indent()
		e.emit(statement)
		e.write(",\n")
	}
	e.depth -= 1
	e.indent()
	e.write(")")
}
