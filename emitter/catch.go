package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitCatchStatement(c *parser.CatchExpression) {
	e.write("try {\n")
	e.depth++
	e.indent()
	e.emit(c.Left)
	e.depth--
	e.write("} catch (")
	e.emitExpression(c.Identifier)
	e.write(") ")
	e.emitBlockStatement(c.Body)
}
