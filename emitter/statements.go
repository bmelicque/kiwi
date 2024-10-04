package emitter

import (
	"fmt"
	"unicode"

	"github.com/bmelicque/test-parser/checker"
)

const maxClassParamsLength = 66

func (e *Emitter) emitAssignment(a checker.Assignment) {
	e.emit(a.Pattern)
	e.write(" = ")
	e.emit(a.Value)
	e.write(";\n")
}

func (e *Emitter) emitBody(b checker.Block) {
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
	e.emit(i.Block)
	if i.Alternate == nil {
		return
	}
	e.write(" else ")
	switch alternate := i.Alternate.(type) {
	case checker.Block:
		e.emitBody(alternate)
	case checker.If:
		e.emitIf(alternate)
	}
}

func (e *Emitter) emitMatchStatement(m checker.MatchExpression) {
	// TODO: break outer loop
	// TODO: declare _m only if calling something
	e.write("const _m = ")
	e.emit(m.Value)
	e.write(";\n")
	if m.Value.Type().Kind() == checker.SUM {
		e.write("switch (_m._tag) {\n")
	} else {
		e.write("switch (_m.constructor) {\n")
	}
	for _, c := range m.Cases {
		e.indent()
		if c.IsCatchall() {
			e.write("default:")
		} else {
			e.write(fmt.Sprintf("case %v: {\n", c.Typing.Text()))
		}
		e.depth++
		if c.Pattern != nil {
			id := c.Pattern.(checker.Identifier)
			e.indent()
			if m.Value.Type().Kind() == checker.SUM {
				e.write(fmt.Sprintf("let %v = _m._value;\n", id.Text()))
			} else {
				e.write(fmt.Sprintf("let %v = _m;\n", id.Text()))
			}
		}
		for _, s := range c.Statements {
			e.indent()
			e.emit(s)
		}
		e.indent()
		e.write("break;\n")
		e.indent()
		e.write("}\n")
		e.depth--
	}
	e.indent()
	e.write("}\n")
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

func (e *Emitter) getClassParamNames(expr checker.Expression) []string {
	params, ok := expr.(checker.TupleExpression)
	if !ok {
		param := expr.(checker.Param)
		return []string{getSanitizedName(param.Identifier.Text())}

	}

	names := make([]string, len(params.Elements))
	length := 0
	for i, member := range params.Elements {
		param := member.(checker.Param)
		name := getSanitizedName(param.Identifier.Text())
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
func (e *Emitter) emitTypeDeclaration(declaration checker.VariableDeclaration) {
	init := declaration.Initializer.Type().(checker.Type).Value.Kind()
	switch init {
	case checker.TRAIT:
		return
	case checker.SUM:
		e.addFlag(SumFlag)
		e.write("class ")
		e.emit(getTypeIdentifier(declaration.Pattern))
		e.write(" extends _Sum {}\n")
		return
	default:
		e.write("class ")
		e.emit(getTypeIdentifier(declaration.Pattern))
		e.write(" {\n    constructor(")
		defer e.write("    }\n}\n")
		object := declaration.Initializer.(checker.ParenthesizedExpression)
		names := e.getClassParamNames(object.Expr)
		e.write(") {\n")
		for _, name := range names {
			e.write(fmt.Sprintf("        this.%v = %v;\n", name, name))
		}
	}
}
func (e *Emitter) emitVariableDeclaration(declaration checker.VariableDeclaration) {
	if isTypePattern(declaration.Pattern) {
		e.emitTypeDeclaration(declaration)
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

func isTypePattern(expr checker.Expression) bool {
	c, ok := expr.(checker.ComputedAccessExpression)
	if ok {
		expr = c
	}
	identifier, ok := expr.(checker.Identifier)
	if !ok {
		return false
	}
	return unicode.IsUpper(rune(identifier.Token.Text()[0]))
}
func getTypeIdentifier(expr checker.Expression) string {
	c, ok := expr.(checker.ComputedAccessExpression)
	if ok {
		expr = c
	}
	identifier := expr.(checker.Identifier)
	return identifier.Text()
}
