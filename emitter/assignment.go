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
		if isMapElementAccess(a.Pattern) || isSliceElement(a.Pattern) {
			emitSetElement(e, a)
		} else {
			emitAssign(e, a.Pattern, a.Value)
		}
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
func emitSetElement(e *Emitter, a *parser.Assignment) {
	pattern := a.Pattern.(*parser.ComputedAccessExpression)
	e.emitExpression(pattern.Expr)
	e.write(".set(")
	e.emitExpression(pattern.Property.Expr)
	e.write(", ")
	e.emitExpression(a.Value)
	e.write(")")
}

func isSliceElement(pattern parser.Expression) bool {
	c, ok := pattern.(*parser.ComputedAccessExpression)
	if !ok {
		return false
	}
	ref, ok := c.Expr.Type().(parser.Ref)
	if !ok {
		return false
	}
	_, ok = ref.To.(parser.List)
	return ok
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
	e.emitBlockStatement(init.Body)
	e.write("\n")
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
