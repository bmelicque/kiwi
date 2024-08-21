package emitter

import (
	"github.com/bmelicque/test-parser/checker"
	"github.com/bmelicque/test-parser/tokenizer"
)

func Precedence(expr checker.Expression) uint8 {
	switch expr := expr.(type) {
	case checker.TupleExpression:
		return 1
	case checker.BinaryExpression:
		switch expr.Operator.Kind() {
		case tokenizer.LOR:
			return 4
		case tokenizer.LAND:
			return 5
		case tokenizer.EQ, tokenizer.NEQ:
			return 9
		case tokenizer.GREATER, tokenizer.GEQ, tokenizer.LESS, tokenizer.LEQ:
			return 10
		case tokenizer.ADD, tokenizer.SUB:
			return 12
		case tokenizer.MUL, tokenizer.DIV, tokenizer.MOD:
			return 13
		case tokenizer.POW:
			return 14
		}
	case checker.CallExpression, checker.PropertyAccessExpression:
		return 18
	case checker.Identifier, checker.Literal, checker.ParenthesizedExpression:
		return 20
	}
	return 0
}
