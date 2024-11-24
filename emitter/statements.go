package emitter

import (
	"fmt"
	"unicode"

	"github.com/bmelicque/test-parser/parser"
)

const maxClassParamsLength = 66

func emitAssign(e *Emitter, pattern parser.Expression, value parser.Expression) {
	e.emit(pattern)
	e.write(" = ")
	switch value.Type().(type) {
	case parser.Nil, parser.Number, parser.Boolean, parser.String, parser.Function:
		e.emit(value)
	default:
		e.write("structuredClone(")
		e.emit(value)
		e.write(")")
	}
	e.write(";\n")
}
func (e *Emitter) emitAssignment(a *parser.Assignment) {
	switch a.Operator.Kind() {
	case parser.Assign:
		if isMapElementAccess(a.Pattern) {
			emitSetMap(e, a)
			return
		}
		emitAssign(e, a.Pattern, a.Value)
	case parser.Declare:
		e.write("let ")
		emitAssign(e, a.Pattern, a.Value)
	case parser.Define:
		if isTypePattern(a.Pattern) {
			e.emitTypeDeclaration(a)
			return
		}
		if _, ok := a.Pattern.(*parser.PropertyAccessExpression); ok {
			e.emitMethodDeclaration(a)
			return
		}

		e.write("const ")
		e.emit(a.Pattern)
		e.write(" = ")
		e.emit(a.Value)
	}
}

func isMapElementAccess(pattern parser.Expression) bool {
	c, ok := pattern.(*parser.ComputedAccessExpression)
	if !ok {
		return false
	}
	alias, ok := c.Expr.Type().(parser.TypeAlias)
	return ok && alias.Name == "Map"
}

// Emit setting a map element, which is not a regular assignment in JS,
// but a call to 'map.set(key, value)'.
func emitSetMap(e *Emitter, a *parser.Assignment) {
	pattern := a.Pattern.(*parser.ComputedAccessExpression)
	e.emitExpression(pattern.Expr)
	e.write(".set(")
	e.emitExpression(pattern.Property.Expr)
	e.write(", ")
	e.emitExpression(a.Value)
	e.write(")")
}

func (e *Emitter) emitBlock(b *parser.Block) {
	e.write("{")
	if len(b.Statements) == 0 {
		e.write("}")
		return
	}
	e.write("\n")
	e.depth++
	for _, statement := range b.Statements {
		e.indent()
		e.emit(statement)
		if _, ok := statement.(parser.Expression); ok {
			e.write(";\n")
		}
	}
	e.depth--
	e.indent()
	e.write("}\n")
}

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
	e.emitBlock(c.Body)
}

func (e *Emitter) emitFor(f *parser.ForExpression) {
	a, ok := f.Statement.(*parser.Assignment)
	if !ok {
		e.write("while (")
		e.emit(f.Statement)
		e.write(") ")
		e.emitBlock(f.Body)
	}

	e.write("for (let ")
	// FIXME: tuples...
	e.emit(a.Pattern)
	e.write(" of ")
	e.emit(a.Value)
	e.write(") ")
	e.emitBlock(f.Body)
}

func (e *Emitter) emitIfStatement(i *parser.IfExpression) {
	e.write("if (")
	e.emit(i.Condition)
	e.write(") ")
	e.emitBlock(i.Body)
	if i.Alternate == nil {
		return
	}
	e.write(" else ")
	switch alternate := i.Alternate.(type) {
	case *parser.Block:
		e.emitBlock(alternate)
	case *parser.IfExpression:
		e.emitIfStatement(alternate)
	}
}

func (e *Emitter) emitMatchStatement(m parser.MatchExpression) {
	// TODO: break outer loop
	// TODO: declare _m only if calling something
	e.write("const _m = ")
	e.emit(m.Value)
	e.write(";\n")
	if _, ok := m.Value.Type().(parser.Sum); ok {
		e.write("switch (_m._tag) {\n")
	} else {
		e.write("switch (_m.constructor) {\n")
	}
	for _, c := range m.Cases {
		e.indent()
		if c.IsCatchall() {
			e.write("default:")
		} else if call, ok := c.Pattern.(*parser.CallExpression); ok {
			e.write("case ")
			e.emit(call.Callee)
			e.write(": {\n")
		} else if id, ok := c.Pattern.(*parser.Identifier); ok {
			e.write("case ")
			e.emit(id)
			e.write(": {\n")
		}
		e.depth++
		if c.Pattern != nil {
			id := c.Pattern.(*parser.Identifier)
			e.indent()
			if _, ok := m.Value.Type().(parser.Sum); ok {
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

func (e *Emitter) emitMethodDeclaration(a *parser.Assignment) {
	pattern := a.Pattern.(*parser.PropertyAccessExpression)
	receiver := pattern.Expr.(*parser.ParenthesizedExpression).Expr.(*parser.Param)

	e.emit(receiver.Complement)
	e.write(".prototype.")
	e.emit(pattern.Property)
	e.write(" = function ")

	e.thisName = receiver.Identifier.Text()
	defer func() { e.thisName = "" }()

	init := a.Value.(*parser.FunctionExpression)
	e.write("(")
	params := init.Params.Expr.(*parser.TupleExpression).Elements
	max := len(params)
	for i := range params[:max] {
		param := params[i].(*parser.Param)
		e.emit(param.Identifier)
		e.write(", ")
	}
	e.emit(params[max].(*parser.Param).Identifier)
	e.write(") ")
	e.emitBlock(init.Body)
	e.write("\n")
}

func (e *Emitter) emitExit(r *parser.Exit) {
	switch r.Operator.Kind() {
	case parser.BreakKeyword:
		e.write("break")
	case parser.ContinueKeyword:
		e.write("continue")
	case parser.ReturnKeyword:
		e.write("return")
	case parser.ThrowKeyword:
		e.write("throw")
	}
	if r.Value != nil {
		e.write(" ")
		e.emit(r.Value)
	}
	e.write(";\n")
}

func (e *Emitter) getClassParamNames(expr parser.Expression) []string {
	params, ok := expr.(*parser.TupleExpression)
	if !ok {
		param := expr.(*parser.Param)
		return []string{getSanitizedName(param.Identifier.Text())}

	}

	names := make([]string, len(params.Elements))
	length := 0
	for i, member := range params.Elements {
		param := member.(*parser.Param)
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
func (e *Emitter) emitTypeDeclaration(declaration *parser.Assignment) {
	switch declaration.Value.Type().(parser.Type).Value.(type) {
	case parser.Trait:
		return
	case parser.Sum:
		e.addFlag(SumFlag)
		e.write("class ")
		e.write(getTypeIdentifier(declaration.Pattern))
		e.write(" extends _Sum {}\n")
		return
	default:
		e.write("class ")
		e.write(getTypeIdentifier(declaration.Pattern))
		e.write(" {\n    constructor(")
		defer e.write("    }\n}\n")
		object := declaration.Value.(*parser.ParenthesizedExpression)
		names := e.getClassParamNames(object.Expr)
		e.write(") {\n")
		for _, name := range names {
			e.write(fmt.Sprintf("        this.%v = %v;\n", name, name))
		}
	}
}

func isTypePattern(expr parser.Expression) bool {
	c, ok := expr.(*parser.ComputedAccessExpression)
	if ok {
		expr = c.Expr
	}
	identifier, ok := expr.(*parser.Identifier)
	if !ok {
		return false
	}
	return unicode.IsUpper(rune(identifier.Token.Text()[0]))
}
func getTypeIdentifier(expr parser.Node) string {
	c, ok := expr.(*parser.ComputedAccessExpression)
	if ok {
		expr = c
	}
	identifier := expr.(*parser.Identifier)
	return identifier.Text()
}
