package parser

import "github.com/bmelicque/test-parser/tokenizer"

type ExpressionTypeKind int

const (
	UNKNOWN ExpressionTypeKind = iota

	TYPE

	NUMBER
	BOOLEAN
	STRING
	NIL

	LIST
	TUPLE
	RANGE

	FUNCTION
)

type ExpressionType interface {
	Kind() ExpressionTypeKind
	Match(ExpressionType) bool
	Extends(ExpressionType) bool
}

type Type struct {
	value ExpressionType
}

func (t Type) Kind() ExpressionTypeKind { return TYPE }
func (t Type) Match(testType ExpressionType) bool {
	return testType.Kind() == TYPE && t.value.Match(testType.(Type).value)
}
func (t Type) Extends(testType ExpressionType) bool {
	return testType.Kind() == TYPE && t.value.Extends(testType.(Type).value)
}

type Primitive struct {
	kind ExpressionTypeKind
}

func (p Primitive) Kind() ExpressionTypeKind    { return p.kind }
func (p Primitive) Match(t ExpressionType) bool { return p.Kind() == t.Kind() }
func (p Primitive) Extends(t ExpressionType) bool {
	return p.Kind() == t.Kind() || p.Kind() == UNKNOWN || t.Kind() == UNKNOWN
}

type List struct {
	element ExpressionType
}

func (l List) Kind() ExpressionTypeKind { return LIST }
func (l List) Match(t ExpressionType) bool {
	if list, ok := t.(List); ok {
		return l.element.Match(list.element)
	}
	return false
}
func (l List) Extends(t ExpressionType) bool {
	if list, ok := t.(List); ok {
		return l.element.Extends(list.element)
	}
	return false
}

type Tuple struct {
	elements []ExpressionType
}

func (t Tuple) Kind() ExpressionTypeKind { return TUPLE }
func (tuple Tuple) Match(t ExpressionType) bool {
	switch t := t.(type) {
	case Tuple:
		if len(t.elements) != len(tuple.elements) {
			return false
		}
		for i := 0; i < len(t.elements); i += 1 {
			if !tuple.elements[i].Match(t.elements[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
func (tuple Tuple) Extends(t ExpressionType) bool {
	switch t := t.(type) {
	case Tuple:
		if len(t.elements) != len(tuple.elements) {
			return false
		}
		for i := 0; i < len(t.elements); i += 1 {
			if !tuple.elements[i].Extends(t.elements[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

type Range struct {
	operands ExpressionType
}

func (r Range) Kind() ExpressionTypeKind { return RANGE }
func (r Range) Match(t ExpressionType) bool {
	if received, ok := t.(Range); ok {
		return r.operands.Match(received.operands)
	}
	return false
}
func (r Range) Extends(t ExpressionType) bool {
	if received, ok := t.(Range); ok {
		return r.operands.Extends(received.operands)
	}
	return false
}

type Function struct {
	params   ExpressionType // Tuple if len > 1
	returned ExpressionType
}

func (f Function) Kind() ExpressionTypeKind      { return FUNCTION }
func (f Function) Match(t ExpressionType) bool   { /* FIXME: */ return false }
func (f Function) Extends(t ExpressionType) bool { /* FIXME: */ return false }

func ReadTypeExpression(expr Expression) ExpressionType {
	switch expr := expr.(type) {
	case TokenExpression:
		switch expr.Token.Kind() {
		case tokenizer.BOOL_KW:
			return Primitive{BOOLEAN}
		case tokenizer.NUM_KW:
			return Primitive{NUMBER}
		case tokenizer.STR_KW:
			return Primitive{STRING}
		}
	}
	return Primitive{UNKNOWN}
}
