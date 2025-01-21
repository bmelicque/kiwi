package parser

import (
	"testing"
)

func TestIsReferencable(t *testing.T) {
	tests := []struct {
		name     string
		expr     Expression
		expected bool
	}{
		{
			name:     "simple identifier",
			expr:     &Identifier{},
			expected: true,
		},
		{
			name:     "type literal", // e.g. &number
			expr:     &Literal{literal{kind: NumberKeyword}},
			expected: true,
		},
		{
			name:     "non-type literal", // e.g. &42
			expr:     &Literal{literal{kind: NumberLiteral}},
			expected: false,
		},
		{
			name: "single property access",
			expr: &PropertyAccessExpression{
				Expr:     &Identifier{},
				Property: &Identifier{},
			},
			expected: true,
		},
		{
			name: "nested property access",
			expr: &PropertyAccessExpression{
				Expr: &PropertyAccessExpression{
					Expr:     &Identifier{}, // object
					Property: &Identifier{}, // nested
				},
				Property: &Identifier{}, // field
			},
			expected: true,
		},
		{
			name: "instance expression",
			expr: &InstanceExpression{
				Typing: &Identifier{},
			},
			expected: true,
		},
		{
			name: "property of instance", // Type{}.key
			expr: &PropertyAccessExpression{
				Expr: &InstanceExpression{
					Typing: &Identifier{},
				},
				Property: &Identifier{},
			},
			expected: false,
		},
		{
			name: "instance of property", // module.Type{}
			expr: &InstanceExpression{
				Typing: &PropertyAccessExpression{
					Expr:     &Identifier{},
					Property: &Identifier{},
				},
			},
			expected: true,
		},
		{
			name:     "nil expression",
			expr:     nil,
			expected: false,
		},
		// Add a test case for an unsupported expression type
		{
			name:     "unsupported expression type",
			expr:     &CatchExpression{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isReferencable(tt.expr)
			if result != tt.expected {
				t.Errorf("isReferencable(%v) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}
