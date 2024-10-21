package emitter

import (
	"github.com/bmelicque/test-parser/parser"
)

func Precedence(expr parser.Expression) uint8 {
	switch expr := expr.(type) {
	case *parser.TupleExpression:
		return 1
	case *parser.BinaryExpression:
		switch expr.Operator.Kind() {
		case parser.LogicalOr:
			return 4
		case parser.LogicalAnd:
			return 5
		case parser.Equal, parser.NotEqual:
			return 9
		case parser.Greater, parser.GreaterEqual, parser.Less, parser.LessEqual:
			return 10
		case parser.Add, parser.Sub:
			return 12
		case parser.Mul, parser.Div, parser.Mod:
			return 13
		case parser.Pow:
			return 14
		}
	case *parser.CallExpression, *parser.PropertyAccessExpression:
		return 18
	case *parser.Identifier, *parser.Literal, *parser.ParenthesizedExpression:
		return 20
	}
	return 0
}
