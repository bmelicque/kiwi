package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

func needsCopy(expr parser.Expression) bool {
	switch expr.Type().(type) {
	case parser.Nil, parser.Number, parser.Boolean, parser.String, parser.Function:
		return false
	}

	switch expr := expr.(type) {
	case *parser.CallExpression,
		*parser.ComputedAccessExpression,
		*parser.PropertyAccessExpression:
		return true
	case *parser.UnaryExpression:
		if expr.Operator.Kind() != parser.Mul {
			return false
		}
		_, isSlice := expr.Type().(parser.List)
		return !isSlice
	}
	return false
}

func emitAssign(e *Emitter, a *parser.Assignment) {
	e.emitExpression(a.Pattern)

	switch a.Operator.Kind() {
	case parser.Assign, parser.Declare:
		e.write(" = ")
	case parser.AddAssign, parser.ConcatAssign:
		e.write(" += ")
	case parser.SubAssign:
		e.write(" -= ")
	case parser.MulAssign:
		e.write(" *= ")
	case parser.DivAssign:
		e.write(" /= ")
	case parser.ModAssign:
		e.write(" %= ")
	case parser.LogicalAndAssign:
		e.write(" &&= ")
	case parser.LogicalOrAssign:
		e.write(" ||= ")
	}

	if needsCopy(a.Value) {
		e.write("structuredClone(")
		e.emitExpression(a.Value)
		e.write(")")
	} else {
		e.emitExpression(a.Value)
	}
	e.write(";\n")
}
func (e *Emitter) emitAssignment(a *parser.Assignment) {
	switch a.Operator.Kind() {
	case parser.Assign:
		if u, ok := a.Pattern.(*parser.UnaryExpression); ok {
			// deref
			e.emitExpression(u.Operand)
			e.write("(0, ")
			e.emitExpression(a.Value)
			e.write(")")
		} else {
			emitAssign(e, a)
		}
	case parser.AddAssign,
		parser.ConcatAssign,
		parser.SubAssign,
		parser.MulAssign,
		parser.DivAssign,
		parser.ModAssign,
		parser.LogicalAndAssign,
		parser.LogicalOrAssign:
		emitAssign(e, a)
	case parser.Declare:
		e.emitDeclaration(a)
	case parser.Define:
		if isTypePattern(a.Pattern) {
			e.emitTypeDeclaration(a)
			return
		}
		if _, ok := a.Pattern.(*parser.PropertyAccessExpression); ok {
			e.emitMethodDeclaration(a)
			return
		}

		if needsExport(a.Pattern) {
			e.write("export ")
		}
		e.write("const ")
		e.emitExpression(a.Pattern)
		e.write(" = ")
		e.emitExpression(a.Value)
	}
}

// FIXME: this is broken: export __sXXX.identifier =
func (e *Emitter) emitDeclaration(a *parser.Assignment) {
	if needsExport(a.Pattern) {
		e.write("export ")
	}
	if i, ok := a.Pattern.(*parser.Identifier); !ok || !isReferenced(i) {
		e.write("let ")
	}
	emitAssign(e, a)
}

func (e *Emitter) emitObjectConstructorParam(n parser.Node) {
	switch n := n.(type) {
	case *parser.Identifier:
		e.emitIdentifier(n)
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
	case *parser.Identifier:
		// parser ensures it is a type identifier,
		// which does not collide with JS reserved names
		name = n.Text()
	case *parser.Param:
		name = getSanitizedName(n.Identifier.Text())
	case *parser.Entry:
		name = getSanitizedName(n.Key.(*parser.Identifier).Text())
	}
	e.indent()
	e.write(fmt.Sprintf("this.%v = %v;\n", name, name))
}

func (e *Emitter) emitObjectTypeDefinition(definition *parser.Assignment) {
	if needsExport(definition.Pattern) {
		e.write("export ")
	}
	b := definition.Value.(*parser.BracedExpression)
	elements := b.Expr.(*parser.TupleExpression).Elements
	if len(elements) == 0 {
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
	max := len(elements) - 1
	for _, s := range elements[:max] {
		e.emitObjectConstructorParam(s)
		e.write(", ")
	}
	e.emitObjectConstructorParam(elements[max])
	e.write(") {\n")

	e.depth++
	for _, s := range elements {
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
	if _, ok := definition.Value.(*parser.BracedExpression); ok {
		e.emitObjectTypeDefinition(definition)
		return
	}
	switch definition.Value.Type().(parser.Type).Value.(type) {
	case parser.Trait:
		return
	case parser.Sum:
		if needsExport(definition.Pattern) {
			e.write("export ")
		}
		e.write("class ")
		e.write(getTypeIdentifier(definition.Pattern))
		e.write(" extends _Sum {}\n")
		return
	}
}

func (e *Emitter) emitMethodDeclaration(a *parser.Assignment) {
	pattern := a.Pattern.(*parser.PropertyAccessExpression)
	receiver := pattern.Expr.(*parser.ParenthesizedExpression).Expr.(*parser.Param)

	e.emitExpression(receiver.Complement)
	e.write(".prototype.")
	e.emitExpression(pattern.Property)
	e.write(" = function ")

	e.thisName = receiver.Identifier.Text()
	defer func() { e.thisName = "" }()

	init := a.Value.(*parser.FunctionExpression)
	e.write("(")
	params := init.Params.Expr.(*parser.TupleExpression).Elements
	max := len(params)
	if max > 0 {
		for i := range params[:max] {
			param := params[i].(*parser.Param)
			e.emitIdentifier(param.Identifier)
			e.write(", ")
		}
		e.emitIdentifier(params[max].(*parser.Param).Identifier)
	}
	e.write(") ")
	e.emitFunctionBody(init.Body, init.Params.Expr.(*parser.TupleExpression))
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
	return identifier.IsType()
}
func getTypeIdentifier(expr parser.Node) string {
	c, ok := expr.(*parser.ComputedAccessExpression)
	if ok {
		expr = c.Expr
	}
	identifier := expr.(*parser.Identifier)
	return identifier.Text()
}

func needsExport(pattern parser.Expression) bool {
	switch pattern := pattern.(type) {
	case *parser.Identifier:
		return !pattern.IsPrivate()
	case *parser.TupleExpression:
		for _, el := range pattern.Elements {
			if needsExport(el) {
				return true
			}
		}
		return false
	case *parser.ComputedAccessExpression:
		return needsExport(pattern.Expr)
	default:
		panic("Case not handled!")
	}
}
