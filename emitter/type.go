package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

func stringify(typing parser.ExpressionType) string {
	switch typing := typing.(type) {
	case parser.Primitive:
		switch typing.Kind() {
		case parser.NUMBER:
			return "number"
		case parser.BOOLEAN:
			return "boolean"
		case parser.STRING:
			return "string"
		}
	case parser.TypeRef:
		return typing.Name
	case parser.List:
		return fmt.Sprintf("%v_list", stringify(typing.Element))
	}
	panic("Not implemented")
}
