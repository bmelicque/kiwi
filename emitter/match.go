package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

func (e *Emitter) emitMatchStatement(m *parser.MatchExpression) {
	e.write("const _m = ")
	e.emitExpression(m.Value)
	e.write(";\n")
	t := m.Value.Type()
	if alias, ok := t.(parser.TypeAlias); ok {
		t = alias.Ref
	}
	if _, ok := t.(parser.Sum); ok {
		emitSumMatch(e, m)
		return
	}
	e.indent()
	e.write("switch (_m.constructor) {\n")
	for _, c := range m.Cases {
		e.indent()
		if c.IsCatchall() {
			e.write("default:")
		} else if call, ok := c.Pattern.(*parser.InstanceExpression); ok {
			e.write("case ")
			e.emitExpression(call.Typing)
			e.write(": {\n")
		} else if id, ok := c.Pattern.(*parser.Identifier); ok {
			e.write("case ")
			e.emitIdentifier(id)
			e.write(": {\n")
		}
		e.depth++
		if c.Pattern != nil {
			pattern := getMatchedName(c.Pattern)
			e.indent()
			e.write(fmt.Sprintf("let %v = _m;\n", pattern))
		}
		e.emitExpression(c.Consequent)
		e.write(";\n")
		e.depth--
		e.indent()
		e.write("}\n")
	}
	e.indent()
	e.write("}\n")
}

func emitSumMatch(e *Emitter, m *parser.MatchExpression) {
	e.indent()
	e.write("switch (_m.tag) {\n")
	for _, c := range m.Cases {
		e.indent()
		if c.IsCatchall() {
			e.write("default:")
		} else if param, ok := c.Pattern.(*parser.Param); ok {
			e.write("case \"")
			e.emitExpression(param.Complement)
			e.write("\": {\n")
		} else if id, ok := c.Pattern.(*parser.Identifier); ok {
			e.write("case \"")
			e.emitIdentifier(id)
			e.write("\": {\n")
		} else {
			panic("unexpected case pattern")
		}
		e.depth++
		if c.Pattern != nil {
			pattern := getMatchedName(c.Pattern)
			e.indent()
			e.write(fmt.Sprintf("let %v = _m.value;\n", pattern))
		}
		e.indent()
		e.emitExpression(c.Consequent)
		e.write(";\n")
		e.depth--
		e.indent()
		e.write("}\n")
	}
	e.indent()
	e.write("}\n")
}

func getMatchedName(pattern parser.Expression) string {
	switch pattern := pattern.(type) {
	case *parser.Identifier:
		return pattern.Text()
	case *parser.Param:
		return pattern.Identifier.Text()
	default:
		panic("unexpected case param")
	}
}
