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
	e.depth--
	e.indent()
	e.write("}\n")
}

func (e *Emitter) emitFunctionExpression(f *parser.FunctionExpression) {
	if f.Type().(parser.Function).Async {
		e.write("async ")
	}
	args := f.Params.Expr.(*parser.TupleExpression).Elements
	e.emitFunctionParams(args)
	e.write(" => ")

	params := f.Params.Expr.(*parser.TupleExpression)
	e.emitFunctionBody(f.Body, params)
}

func (e *Emitter) emitFunctionParams(params []parser.Expression) {
	if len(params) == 0 {
		e.write("()")
		return
	}

	e.write("(")
	max := len(params) - 1
	for i := range params[:max] {
		e.emitFunctionParam(params[i])
		e.write(", ")
	}
	e.emitFunctionParam(params[max])
	e.write(")")
}

func (e *Emitter) emitFunctionParam(arg parser.Expression) {
	switch arg := arg.(type) {
	case *parser.Param:
		e.emitIdentifier(arg.Identifier)
	case *parser.Identifier:
		e.emitIdentifier(arg)
	default:
		panic("expected param or identifier")
	}
}
