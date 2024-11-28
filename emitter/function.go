package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

func (e *Emitter) emitFunctionBody(b *parser.Block, params *parser.TupleExpression) {
	e.write("{")
	if len(b.Statements) == 0 {
		e.write("}")
		return
	}
	e.write("\n")
	e.depth++

	for _, param := range params.Elements {
		if _, ok := param.Type().(parser.Ref); ok {
			continue
		}
		name := param.(*parser.Param).Identifier.Text()
		v, ok := b.Scope().Find(name)
		if !ok {
			panic("variable should be found in current scope...")
		}
		if isMutated(v) {
			e.indent()
			e.write(fmt.Sprintf("%v = structuredClone(%v);\n", name, name))
		}
	}
	max := len(b.Statements) - 1
	for _, statement := range b.Statements[:max] {
		e.indent()
		e.emit(statement)
	}
	e.indent()
	e.write("return ")
	e.emit(b.Statements[max])
	e.write(";\n")
	e.depth--
	e.indent()
	e.write("}\n")
}

func (e *Emitter) emitFunctionExpression(f *parser.FunctionExpression) {
	if f.Type().(parser.Function).Async {
		e.write("async ")
	}
	e.write("(")
	args := f.Params.Expr.(*parser.TupleExpression).Elements
	max := len(args) - 1
	for i := range args[:max] {
		param := args[i].(*parser.Param)
		e.emit(param.Identifier)
		e.write(", ")
	}
	e.emit(args[max].(*parser.Param).Identifier)
	e.write(") => ")

	params := f.Params.Expr.(*parser.TupleExpression)
	e.emitFunctionBody(f.Body, params)
}
