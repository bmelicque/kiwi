package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestCatchStatement(t *testing.T) {
	emitter := makeEmitter()
	// result catch _ { 0 }
	emitter.emit(&parser.CatchExpression{
		Left:       &parser.Identifier{Token: testToken{kind: parser.Name, value: "result"}},
		Identifier: &parser.Identifier{Token: testToken{kind: parser.Name, value: "_"}},
		Body: &parser.Block{Statements: []parser.Node{
			&parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "0"}},
		}},
	})

	text := emitter.string()
	expected := "try {\n"
	expected += "    result;\n"
	expected += "} catch (_) {\n"
	expected += "    0;\n"
	expected += "}\n"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

func TestMapAssignment(t *testing.T) {
	emitter := makeEmitter()
	// map[key] = value
	emitSetMap(emitter, &parser.Assignment{
		Pattern: &parser.ComputedAccessExpression{
			Expr: &parser.Identifier{Token: testToken{kind: parser.Name, value: "map"}},
			Property: &parser.BracketedExpression{
				Expr: &parser.Literal{Token: testToken{kind: parser.StringLiteral, value: "\"key\""}},
			},
		},
		Value:    &parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "42"}},
		Operator: testToken{kind: parser.Assign},
	})

	text := emitter.string()
	expected := "map.set(\"key\", 42)"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

func TestIfStatement(t *testing.T) {
	emitter := makeEmitter()
	emitter.emit(&parser.IfExpression{
		Condition: &parser.Literal{
			Token: testToken{kind: parser.BooleanLiteral, value: "false"},
		},
		Body: &parser.Block{Statements: []parser.Node{}},
		Alternate: &parser.IfExpression{
			Condition: &parser.Literal{
				Token: testToken{kind: parser.BooleanLiteral, value: "false"},
			},
			Body:      &parser.Block{Statements: []parser.Node{}},
			Alternate: &parser.Block{},
		},
	})

	text := emitter.string()
	expected := "if (false) {} else if (false) {} else {}"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}
