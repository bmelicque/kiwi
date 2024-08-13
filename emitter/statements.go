package emitter

import (
	"fmt"
	"unicode"

	"github.com/bmelicque/test-parser/checker"
)

const maxClassParamsLength = 66

func (e *Emitter) emitClassParams(params []checker.ObjectMemberDefinition) []string {
	names := make([]string, len(params))
	length := 0
	for i, member := range params {
		name := member.Name.Token.Text()
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

func (e *Emitter) emitClass(declaration checker.VariableDeclaration) {
	e.Write("class ")
	e.Emit(declaration.Pattern)
	e.Write(" {\n    constructor(")
	defer e.Write("    }\n}\n")

	names := e.emitClassParams(declaration.Initializer.(checker.ObjectDefinition).Members)
	e.Write(") {\n")
	for _, name := range names {
		e.Write(fmt.Sprintf("        this.%v = %v;\n", name, name))
	}
}

func (e *Emitter) emitAssignment(a checker.Assignment) {
	e.Emit(a.Pattern)
	e.Write(" = ")
	e.Emit(a.Value)
	e.Write(";\n")
}

func (e *Emitter) emitBody(b checker.Body) {
	e.Write("{")
	if len(b.Statements) == 0 {
		e.Write("}")
		return
	}
	e.Write("\n")
	defer func() {
		e.Indent()
		e.Write("}\n")
	}()
	e.depth += 1
	defer func() { e.depth -= 1 }()
	for _, statement := range b.Statements {
		e.Indent()
		e.Emit(statement)
	}
}

func (e *Emitter) emitExpressionStatement(s checker.ExpressionStatement) {
	e.Emit(s.Expr)
	e.Write(";\n")
}

func (e *Emitter) emitFor(f checker.For) {
	e.Write("while (")
	e.Emit(f.Condition)
	e.Write(") ")
	e.Emit(f.Body)
}

func (e *Emitter) emitForRange(f checker.ForRange) {
	e.Write("for (")
	if f.Declaration.Constant {
		e.Write("const")
	} else {
		e.Write("let")
	}
	e.Write(" ")
	// FIXME: tuples...
	e.Emit(f.Declaration.Pattern)
	e.Write(" of ")
	e.Emit(f.Declaration.Range)
	e.Write(") ")
	e.Emit(f.Body)
}

func (e *Emitter) emitIf(i checker.If) {
	e.Write("if (")
	e.Emit(i.Condition)
	e.Write(") ")
	e.Emit(i.Body)
}

func (e *Emitter) emitMethodDeclaration(method checker.MethodDeclaration) {
	e.Emit(method.Receiver.Typing)
	e.Write(".prototype.")
	e.Emit(method.Name)
	e.Write(" = function ")

	e.thisName = method.Receiver.Name.Text()
	defer func() { e.thisName = "" }()

	switch init := method.Initializer.(type) {
	case checker.FatArrowFunction:
		e.emitParams(init.Params)
		e.Write(" ")
		e.emitBody(init.Body)
	case checker.SlimArrowFunction:
		e.emitParams(init.Params)
		e.Write(" { return ")
		e.Emit(init.Expr)
		e.Write(" }")
	}
	e.Write("\n")
}

func (e *Emitter) emitReturn(r checker.Return) {
	e.Write("return")
	if r.Value != nil {
		e.Write(" ")
		e.Emit(r.Value)
	}
	e.Write(";\n")
}

func (e *Emitter) emitVariableDeclaration(declaration checker.VariableDeclaration) {
	if isTypeIdentifier(declaration.Pattern) {
		e.emitClass(declaration)
		return
	}

	if declaration.Constant {
		e.Write("const ")
	} else {
		e.Write("let ")
	}

	e.Emit(declaration.Pattern)
	e.Write(" = ")
	e.Emit(declaration.Initializer)
	e.Write(";\n")
}

func isTypeIdentifier(expr checker.Expression) bool {
	identifier, ok := expr.(checker.Identifier)
	if !ok {
		return false
	}
	return unicode.IsUpper(rune(identifier.Token.Text()[0]))
}
