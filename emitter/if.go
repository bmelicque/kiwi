package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitIfStatement(i *parser.IfExpression) {
	e.write("if (")
	e.emit(i.Condition)
	e.write(") ")
	e.emitBlockStatement(i.Body)
	if i.Alternate == nil {
		return
	}
	e.write(" else ")
	switch alternate := i.Alternate.(type) {
	case *parser.Block:
		e.emitBlockStatement(alternate)
	case *parser.IfExpression:
		e.emitIfStatement(alternate)
	}
}

func (e *Emitter) emitIfExpression(i *parser.IfExpression) {
	e.emit(i.Condition)
	e.write(" ? ")
	e.emitBlockExpression(i.Body)
	e.write(" : ")
	if i.Alternate == nil {
		e.write("undefined")
		return
	}
	switch alternate := i.Alternate.(type) {
	case *parser.Block:
		e.emitBlockExpression(alternate)
	case *parser.IfExpression:
		e.emitIfExpression(alternate)
	}
}
