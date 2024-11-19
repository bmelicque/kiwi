package parser

import (
	"strings"
	"testing"
)

func TestParseSumType(t *testing.T) {
	str := "| Some{Type} | None"
	parser, err := MakeParser(strings.NewReader(str))
	if err != nil {
		t.Fatal(err)
	}
	node := parser.parseSumType()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}

	sum, ok := node.(*SumType)
	if !ok {
		t.Fatalf("Expected SumType, got %#v", node)
		return
	}
	if len(sum.Members) != 2 {
		t.Fatalf("Expected 2 elements, got %v: %#v", len(sum.Members), sum.Members)
	}
}

func TestSumTypeLength(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("| Alone"))
	if err != nil {
		t.Fatal(err)
	}
	parser.parseSumType()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestCheckSumType(t *testing.T) {
	parser, _ := MakeParser(nil)
	expr := &SumType{Members: []SumTypeConstructor{
		{
			Name: &Identifier{Token: literal{kind: Name, value: "A"}},
			Params: &BracedExpression{
				Expr: &TupleExpression{Elements: []Expression{
					&Literal{token{kind: NumberKeyword}},
				}},
			},
		},
		{
			Name: &Identifier{Token: literal{kind: Name, value: "B"}},
		},
	}}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
	if _, ok := expr.Type().(Type); !ok {
		t.Fatalf("Expected type, got %v", expr)
	}
}
