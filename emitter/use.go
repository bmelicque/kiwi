package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitUseStatement(u *parser.UseDirective) {
	path := u.Source.Text()
	path = path[1 : len(path)-1]
	if parser.IsLocalPath(path) {
		e.emitLocalImport(u)
		return
	}
	switch path {
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
	e.write("\n")
}
