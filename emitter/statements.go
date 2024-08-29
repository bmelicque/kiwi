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
		e.write("\n")
		for _, name := range names {
			e.write("        ")
			e.write(name)
			e.write(",\n")
		}
		e.write("    ")
	} else {
		for i, name := range names {
			e.write(name)
			if i != len(names)-1 {
				e.write(", ")
			}
		}
	}
	return names
}

func (e *Emitter) emitClass(declaration checker.VariableDeclaration) {
	e.write("class ")
	e.emit(declaration.Pattern)
	e.write(" {\n    constructor(")
	defer e.write("    }\n}\n")

	object := declaration.Initializer.(checker.ObjectDefinition)
	if generic, ok := declaration.Initializer.(checker.GenericTypeDef); ok {
		object = generic.Expr.(checker.ObjectDefinition)
	}
	names := e.emitClassParams(object.Members)
	e.write(") {\n")
	for _, name := range names {
		e.write(fmt.Sprintf("        this.%v = %v;\n", name, name))
	}
}

func (e *Emitter) emitAssignment(a checker.Assignment) {
	e.emit(a.Pattern)
	e.write(" = ")
	e.emit(a.Value)
	e.write(";\n")
}

func (e *Emitter) emitBody(b checker.Body) {
	e.write("{")
	if len(b.Statements) == 0 {
		e.write("}")
		return
	}
	e.write("\n")
	defer func() {
		e.indent()
		e.write("}\n")
	}()
	e.depth += 1
	defer func() { e.depth -= 1 }()
	for _, statement := range b.Statements {
		e.indent()
		e.emit(statement)
	}
}

func (e *Emitter) emitExpressionStatement(s checker.ExpressionStatement) {
	e.emit(s.Expr)
	e.write(";\n")
}

func (e *Emitter) emitFor(f checker.For) {
	e.write("while (")
	e.emit(f.Condition)
	e.write(") ")
	e.emit(f.Body)
}

func (e *Emitter) emitForRange(f checker.ForRange) {
	e.write("for (")
	if f.Declaration.Constant {
		e.write("const")
	} else {
		e.write("let")
	}
	e.write(" ")
	// FIXME: tuples...
	e.emit(f.Declaration.Pattern)
	e.write(" of ")
	e.emit(f.Declaration.Range)
	e.write(") ")
	e.emit(f.Body)
}

func (e *Emitter) emitIf(i checker.If) {
	e.write("if (")
	e.emit(i.Condition)
	e.write(") ")
	e.emit(i.Body)
}

func (e *Emitter) emitMethodDeclaration(method checker.MethodDeclaration) {
	e.emit(method.Receiver.Typing)
	e.write(".prototype.")
	e.emit(method.Name)
	e.write(" = function ")

	e.thisName = method.Receiver.Name.Text()
	defer func() { e.thisName = "" }()

	switch init := method.Initializer.(type) {
	case checker.FatArrowFunction:
		e.emitParams(init.Params)
		e.write(" ")
		e.emitBody(init.Body)
	case checker.SlimArrowFunction:
		e.emitParams(init.Params)
		e.write(" { return ")
		e.emit(init.Expr)
		e.write(" }")
	}
	e.write("\n")
}

func (e *Emitter) emitReturn(r checker.Return) {
	e.write("return")
	if r.Value != nil {
		e.write(" ")
		e.emit(r.Value)
	}
	e.write(";\n")
}

func (e *Emitter) emitVariableDeclaration(declaration checker.VariableDeclaration) {
	if isTypeIdentifier(declaration.Pattern) {
		e.emitClass(declaration)
		return
	}

	if declaration.Constant {
		e.write("const ")
	} else {
		e.write("let ")
	}

	e.emit(declaration.Pattern)
	e.write(" = ")
	e.emit(declaration.Initializer)
	switch declaration.Initializer.(type) {
	case checker.SlimArrowFunction, checker.FatArrowFunction:
	default:
		e.write(";\n")
	}
}

func isTypeIdentifier(expr checker.Expression) bool {
	identifier, ok := expr.(checker.Identifier)
	if !ok {
		return false
	}
	return unicode.IsUpper(rune(identifier.Token.Text()[0]))
}
