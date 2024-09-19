package emitter

import "slices"

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
