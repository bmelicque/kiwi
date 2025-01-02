package emitter

import "github.com/bmelicque/test-parser/parser"

func (e *Emitter) emitUseStatement(u *parser.UseDirective) {
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
