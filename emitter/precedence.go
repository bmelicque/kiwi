package emitter

import (
	"github.com/bmelicque/test-parser/checker"
	"github.com/bmelicque/test-parser/parser"
)

func Precedence(expr checker.Expression) uint8 {
	switch expr := expr.(type) {
	case checker.TupleExpression:
		return 1
	case checker.BinaryExpression:
		switch expr.Operator.Kind() {
		case parser.LOR:
			return 4
		case parser.LAND:
			return 5
		case parser.EQ, parser.NEQ:
			return 9
		case parser.GREATER, parser.GEQ, parser.LESS, parser.LEQ:
			return 10
		case parser.ADD, parser.SUB:
			return 12
		case parser.MUL, parser.DIV, parser.MOD:
			return 13
		case parser.POW:
			return 14
		}
	case checker.CallExpression, checker.PropertyAccessExpression:
		return 18
	case checker.Identifier, checker.Literal, checker.ParenthesizedExpression:
		return 20
	}
	return 0
}
