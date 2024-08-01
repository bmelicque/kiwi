package emitter

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func (e *Emitter) EmitAssignment(a parser.Assignment) {
	if parser.IsType(a.Declared) {
		return
	}

	kind := a.Operator.Kind()
	if kind == tokenizer.DEFINE {
		e.Write("const ")
	} else if kind == tokenizer.DECLARE || kind == tokenizer.ASSIGN && a.Typing != nil {
		e.Write("let ")
	}
	method, ok := a.Declared.(parser.PropertyAccessExpression)
	if ok {
		e.Emit(method.Expr.(parser.TupleExpression).Elements[0].(parser.TypedExpression).Typing)
		e.Write("_")
		e.Emit(method.Property)
	} else {
		e.Emit(a.Declared)
	}
	e.Write(" = ")
	if ok {
		e.emitMethod(a.Initializer.(parser.FunctionExpression), method.Expr.(parser.TupleExpression).Elements[0].(parser.TypedExpression))
	} else {
		e.Emit(a.Initializer)
	}
	if _, ok := a.Initializer.(parser.FunctionExpression); !ok {
		e.Write(";")
	}
	e.Write("\n")
}

func (e *Emitter) EmitBody(b parser.Body) {
	e.Write("{")
	if len(b.Statements) == 0 {
		e.Write("}")
		return
	}
	e.Write("\n")

	e.depth += 1
	for _, statement := range b.Statements {
		e.Indent()
		e.Emit(statement)
	}
	e.depth -= 1

	e.Indent()
	e.Write("}\n")
}

func (e *Emitter) EmitExpressionStatement(s parser.ExpressionStatement) {
	e.Emit(s.Expr)
	e.Write(";\n")
}

func (e *Emitter) EmitFor(f parser.For) {
	if assignment, ok := f.Statement.(parser.Assignment); ok {
		e.Write("for (const ")
		e.Emit(assignment.Declared)
		e.Write(" of ")
		e.Emit(assignment.Initializer)
	} else {
		e.Write("while (")
		if f.Statement != nil {
			// FIXME: ';' at the end of the statement, body should handle where to put ';'
			e.Emit(f.Statement)
		} else {
			e.Write("true")
		}
	}
	e.Write(") ")

	e.Emit(*f.Body)
}

// TODO: handle alternate
func (e *Emitter) EmitIfElse(i parser.IfElse) {
	e.Write("if (")
	e.Emit(i.Condition)
	e.Write(")")

	e.Write(" ")
	e.Emit(*i.Body)
}

func (e *Emitter) EmitReturn(r parser.Return) {
	e.Write("return")
	if r.Value != nil {
		e.Write(" ")
		e.Emit(r.Value)
	}
	e.Write(";\n")
}
