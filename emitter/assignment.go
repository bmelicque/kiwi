package emitter

import (
	"fmt"
	"unicode"

	"github.com/bmelicque/test-parser/parser"
)

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

func (e *Emitter) emitObjectConstructorParam(n parser.Node) {
	switch n := n.(type) {
	case *parser.Param:
		e.emitIdentifier(n.Identifier)
	case *parser.Entry:
		e.emitIdentifier(n.Key.(*parser.Identifier))
		e.write(" = ")
		e.emitExpression(n.Value)
	}
}

func (e *Emitter) emitObjectConstructorStatement(n parser.Node) {
	var name string
	switch n := n.(type) {
	case *parser.Param:
		name = getSanitizedName(n.Identifier.Text())
	case *parser.Entry:
		name = getSanitizedName(n.Key.(*parser.Identifier).Text())
	}
	e.indent()
	e.write(fmt.Sprintf("this.%v = %v;\n", name, name))
}

func (e *Emitter) emitObjectTypeDefinition(definition *parser.Assignment) {
	b := definition.Value.(*parser.Block)
	if len(b.Statements) == 0 {
		e.write("class ")
		e.write(getTypeIdentifier(definition.Pattern))
		e.write(" {}\n")
		return
	}

	e.write("class ")
	e.write(getTypeIdentifier(definition.Pattern))
	e.write(" {\n")

	e.depth++
	e.indent()
	e.write("constructor(")
	max := len(b.Statements) - 1
	for _, s := range b.Statements[:max] {
		e.emitObjectConstructorParam(s)
		e.write(", ")
	}
	e.emitObjectConstructorParam(b.Statements[max])
	e.write(") {\n")

	e.depth++
	for _, s := range b.Statements {
		e.emitObjectConstructorStatement(s)
	}
	e.depth--
	e.indent()
	e.write("}\n")
	e.depth--
	e.indent()
	e.write("}\n")
}

func (e *Emitter) emitTypeDeclaration(definition *parser.Assignment) {
	if _, ok := definition.Value.(*parser.Block); ok {
		e.emitObjectTypeDefinition(definition)
		return
	}
	switch definition.Value.Type().(parser.Type).Value.(type) {
	case parser.Trait:
		return
	case parser.Sum:
		e.addFlag(SumFlag)
		e.write("class ")
		e.write(getTypeIdentifier(definition.Pattern))
		e.write(" extends _Sum {}\n")
		return
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
		expr = c.Expr
	}
	identifier := expr.(*parser.Identifier)
	return identifier.Text()
}
