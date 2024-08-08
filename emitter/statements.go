package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

const maxClassParamsLength = 66

func (e *Emitter) emitParams(params []parser.Node) []string {
	names := make([]string, len(params))
	length := 0
	for i, member := range params {
		name := member.(parser.TypedExpression).Expr.(*parser.TokenExpression).Token.Text()
		names[i] = name
		length += len(name) + 2
	}

	if length > maxClassParamsLength {
		e.Write("\n")
		for _, name := range names {
			e.Write("        ")
			e.Write(name)
			e.Write(",\n")
		}
		e.Write("    ")
	} else {
		for i, name := range names {
			e.Write(name)
			if i != len(names)-1 {
				e.Write(", ")
			}
		}
	}
	return names
}

func (e *Emitter) emitClass(a parser.Assignment) {
	e.Write("class ")
	e.Emit(a.Declared)
	e.Write(" {\n    constructor(")
	defer e.Write("    }\n}\n")

	names := e.emitParams(a.Initializer.(parser.ObjectDefinition).Members)
	e.Write(") {\n")
	for _, name := range names {
		e.Write(fmt.Sprintf("        this.%v = %v;\n", name, name))
	}
}

func (e *Emitter) emitMethodDeclaration(method *parser.PropertyAccessExpression, function parser.FunctionExpression) {
	expr := method.Expr.(parser.TupleExpression).Elements[0].(parser.TypedExpression)
	e.Emit(expr.Typing)
	e.Write(".prototype.")
	e.Emit(method.Property)
	e.Write(" = function (")
	e.emitParams(function.Params.Elements)
	e.Write(") ")

	e.thisName = expr.Expr.(*parser.TokenExpression).Token.Text()
	defer func() { e.thisName = "" }()
	if function.Body != nil {
		e.EmitBody(*function.Body)
	} else {
		e.Write("{ return ")
		e.Emit(function.Expr)
		e.Write(" }")
	}
	e.Write("\n")
}

func (e *Emitter) EmitAssignment(a parser.Assignment) {
	if parser.IsTypeToken(a.Declared) {
		e.emitClass(a)
		return
	}

	if method, ok := a.Declared.(*parser.PropertyAccessExpression); ok {
		e.emitMethodDeclaration(method, a.Initializer.(parser.FunctionExpression))
		return
	}

	kind := a.Operator.Kind()
	if kind == tokenizer.DEFINE {
		e.Write("const ")
	} else if kind == tokenizer.DECLARE || kind == tokenizer.ASSIGN && a.Typing != nil {
		e.Write("let ")
	}

	e.Emit(a.Declared)
	e.Write(" = ")
	e.Emit(a.Initializer)

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
