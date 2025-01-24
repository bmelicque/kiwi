package emitter

import (
	"fmt"
	"strings"

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
		emitCaseStatements(e, c.Statements)
		e.depth--
		e.indent()
		e.write("}\n")
	}
	e.indent()
	e.write("}\n")
}

func emitSumMatch(e *Emitter, m *parser.MatchExpression) {
	e.write("switch (_m.tag) {\n")
	for _, c := range m.Cases {
		e.indent()
		if c.IsCatchall() {
			e.write("default:")
		} else if call, ok := c.Pattern.(*parser.InstanceExpression); ok {
			e.write("case \"")
			e.emitExpression(call.Typing)
			e.write("\": {\n")
		} else if id, ok := c.Pattern.(*parser.Identifier); ok {
			e.write("case \"")
			e.emitIdentifier(id)
			e.write("\": {\n")
		}
		e.depth++
		if c.Pattern != nil {
			pattern := getMatchedName(c.Pattern)
			e.indent()
			e.write(fmt.Sprintf("let %v = _m.value;\n", pattern))
		}
		emitCaseStatements(e, c.Statements)
		e.indent()
		e.write("}\n")
	}
	e.indent()
	e.write("}\n")
}

func emitCaseStatements(e *Emitter, statements []parser.Node) {
	for _, s := range statements {
		e.indent()
		e.emit(s)
	}
	e.indent()
	e.write("break;\n")
	e.depth--
}

func getMatchedName(pattern parser.Expression) string {
	switch pattern := pattern.(type) {
	case *parser.Identifier:
		return pattern.Text()
	case *parser.InstanceExpression:
		elements := parser.MakeTuple(pattern.Args.Expr).Elements
		last := len(elements) - 1
		builder := strings.Builder{}
		for _, element := range elements[:last] {
			id := element.(*parser.Identifier)
			builder.WriteString(id.Text())
			builder.WriteString(", ")
		}
		builder.WriteString(elements[last].(*parser.Identifier).Text())
		return builder.String()
	default:
		panic("unexpected case param")
	}
}
