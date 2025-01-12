package emitter

import (
	"slices"

	"github.com/bmelicque/test-parser/parser"
)

func (e *Emitter) emitUseStatement(u *parser.UseDirective) {
	path := u.Source.Text()
	path = path[1 : len(path)-1]
	if parser.IsLocalPath(path) {
		e.emitLocalImport(u)
		return
	}
	switch path {
	case "dom":
		if u.Star {
			e.write("const ")
			e.emitExpression(u.Names)
			e.write(" = { createElement: (s) => document.createElement(s) }\n")
			return
		}
		names := getUsedNames(u.Names)
		if slices.Contains(names, "createElement") {
			e.write("const createElement = (s) => document.createElement(s);\n")
		}
	case "io":
		if u.Star {
			e.write("const ")
			e.emitExpression(u.Names)
			e.write(" = console;\n")
		} else {
			e.write("const {")
			e.emitExpression(u.Names)
			e.write("} = console;\n")
		}
	}
}

func getUsedNames(n parser.Expression) []string {
	switch n := n.(type) {
	case *parser.Identifier:
		return []string{n.Text()}
	case *parser.TupleExpression:
		names := make([]string, len(n.Elements))
		for i := range n.Elements {
			names[i] = n.Elements[i].(*parser.Identifier).Text()
		}
		return names
	default:
		panic("unexpected names")
	}
}

func (e *Emitter) emitLocalImport(u *parser.UseDirective) {
	e.write("import ")
	if u.Star {
		e.write("* as ")
		e.emitExpression(u.Names)
	} else {
		e.write("{")
		e.emitExpression(u.Names)
		e.write("} ")
	}
	e.write("from ")
	e.emitExpression(u.Source)
	e.write(".js\n")
}
