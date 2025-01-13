package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

func (e *Emitter) emitUnaryExpression(u *parser.UnaryExpression) {
	switch u.Operator.Kind() {
	case parser.AsyncKeyword:
		e.emitCallExpression(u.Operand.(*parser.CallExpression), false)
	case parser.AwaitKeyword:
		e.write("await ")
		e.emitExpression(u.Operand)
	case parser.Bang:
		e.write("!")
		e.emitExpression(u.Operand)
	case parser.TryKeyword:
		e.emitExpression(u.Operand)
	case parser.BinaryAnd:
		e.emitReference(u.Operand)
	case parser.Mul:
		e.emitExpression(u.Operand)
		if _, ok := u.Operand.Type().(parser.Ref).To.(parser.List); ok {
			e.write(".clone()")
		} else {
			e.write("(1)")
		}
	}
}

func (e *Emitter) emitReference(expr parser.Expression) {
	if implementsNode(expr.Type()) {
		e.write("new __.NodePointer(")
	} else {
		e.write("new __.Pointer(")
	}
	identifier, isIdentifier := getRefIdentifier(expr)
	if isIdentifier {
		emitScope(e, identifier.GetScope())
	} else {
		e.emitExpression(expr.(*parser.PropertyAccessExpression).Expr)
	}
	e.write(fmt.Sprintf(", \"%v\")", identifier.Text()))
}
func getRefIdentifier(expr parser.Expression) (*parser.Identifier, bool) {
	switch expr := expr.(type) {
	case *parser.Identifier:
		return expr, true
	case *parser.PropertyAccessExpression:
		return expr.Property.(*parser.Identifier), false
	default:
		panic("unexpected ref expression (should be &identifier or &object.prop)")
	}
}

func isReferenceExpression(expr parser.Expression) bool {
	u, ok := expr.(*parser.UnaryExpression)
	return ok && u.Operator.Kind() == parser.BinaryAnd
}
