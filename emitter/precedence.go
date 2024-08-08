package emitter

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func Precedence(expr parser.Node) int8 {
	switch expr := expr.(type) {
	case parser.BinaryExpression:
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
	case parser.CallExpression, *parser.PropertyAccessExpression:
		return 18
	case parser.TupleExpression:
		if len(expr.Elements) > 1 {
			return 19
		} else {
			return Precedence(expr.Elements[0])
		}
	case *parser.TokenExpression:
		return 20
	}
	return 0
}
