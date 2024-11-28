package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitCatchStatement(c *parser.CatchExpression) {
	e.write("try {\n")
	e.depth++
	e.indent()
	e.emit(c.Left)
	e.write(";\n")
	e.depth--
	e.write("} catch (")
	e.emit(c.Identifier)
	e.write(") ")
	e.emitBlockStatement(c.Body)
}
