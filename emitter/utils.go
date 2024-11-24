package emitter

import (
	"slices"

	"github.com/bmelicque/test-parser/parser"
)

var reservedWords = []string{
	"abstract", "arguments", "boolean", "byte", "case", "char", "class",
	"const", "debugger", "default", "delete", "do", "double", "enum", "eval",
	"export", "extends", "final", "finally", "float", "function", "goto",
	"implements", "import", "in", "instanceof", "int", "interface", "let",
	"long", "native", "new", "null", "package", "private", "protected",
	"public", "short", "static", "super", "switch", "synchronized", "this",
	"throw", "throws", "transient", "typeof", "var", "void", "volatile",
	"while", "with", "yield",
}

func getSanitizedName(name string) string {
	if slices.Contains(reservedWords, name) {
		return name + "_"
	}
	return name
}

// Check if a variable is mutated. Re-assigns are not accounted for.
// e.g. `object.key = value` is listed, not `variable = value`
func isMutated(v *parser.Variable) bool {
	switch v.Typing.(type) {
	case parser.Boolean, parser.Nil, parser.Number, parser.String:
		return false
	}
	writes := v.Writes()
	for _, write := range writes {
		switch write := write.(type) {
		case *parser.Assignment:
			if _, ok := write.Pattern.(*parser.Identifier); ok {
				return true
			}
		case *parser.UnaryExpression:
			return true
		default:
			panic("Invalid type for writes")
		}
	}
	return false
}
