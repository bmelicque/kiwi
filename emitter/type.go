package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

func stringify(typing parser.ExpressionType) string {
	switch typing := typing.(type) {
	case parser.Primitive:
		// TODO:
	case parser.TypeRef:
		return typing.Name
	case parser.List:
		return fmt.Sprintf("%v_list", stringify(typing.Element))
	}
	panic("Not implemented")
}
