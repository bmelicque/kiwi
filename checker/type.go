package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type ExpressionTypeKind int

const (
	UNKNOWN ExpressionTypeKind = iota

	TYPE

	NUMBER
	BOOLEAN
	STRING
	NIL

	TYPE_REF

	LIST
	TUPLE
	RANGE
	STRUCT

	FUNCTION
)

type ExpressionType interface {
	Kind() ExpressionTypeKind
	Match(ExpressionType) bool
	Extends(ExpressionType) bool
	build(*Scope) ExpressionType
}

type Type struct {
	Value ExpressionType
}

func (t Type) Kind() ExpressionTypeKind { return TYPE }
func (t Type) Match(testType ExpressionType) bool {
	return testType.Kind() == TYPE && t.Value.Match(testType.(Type).Value)
}
func (t Type) Extends(testType ExpressionType) bool {
	return testType.Kind() == TYPE && t.Value.Extends(testType.(Type).Value)
}
func (t Type) build(scope *Scope) ExpressionType { return Type{t.Value.build(scope)} }

type Primitive struct {
	kind ExpressionTypeKind
}

func (p Primitive) Kind() ExpressionTypeKind    { return p.kind }
func (p Primitive) Match(t ExpressionType) bool { return p.Kind() == t.Kind() }
func (p Primitive) Extends(t ExpressionType) bool {
	if t == nil {
		return true
	}
	return p.Kind() == t.Kind() || p.Kind() == UNKNOWN || t.Kind() == UNKNOWN
}
func (p Primitive) build(scope *Scope) ExpressionType { return p }

type TypeAlias struct {
	Name   string
	Params []ExpressionType
	Ref    ExpressionType
}

func (t TypeAlias) Kind() ExpressionTypeKind { return t.Ref.Kind() }
func (ta TypeAlias) Match(t ExpressionType) bool {
	alias, ok := t.(TypeAlias)
	if !ok {
		return false
	}
	if alias.Name != ta.Name {
		return false
	}
	for i, param := range ta.Params {
		if param != nil && !param.Match(alias.Params[i]) {
			return false
		}
	}
	return true
}
func (ta TypeAlias) Extends(t ExpressionType) bool {
	alias, ok := t.(TypeAlias)
	if !ok {
		return false
	}
	if alias.Name != ta.Name && !ta.Ref.Extends(alias.Ref) {
		return false
	}
	for i, param := range ta.Params {
		if param != nil && !param.Extends(alias.Params[i]) {
			return false
		}
	}
	return true
}

func (ta TypeAlias) build(scope *Scope) ExpressionType {
	ta.Ref = ta.Ref.build(scope)
	return ta
}

type List struct {
	Element ExpressionType
}

func (l List) Kind() ExpressionTypeKind { return LIST }
func (l List) Match(t ExpressionType) bool {
	if list, ok := t.(List); ok {
		return l.Element.Match(list.Element)
	}
	return false
}
func (l List) Extends(t ExpressionType) bool {
	if list, ok := t.(List); ok {
		return l.Element.Extends(list.Element)
	}
	return false
}
func (l List) build(scope *Scope) ExpressionType {
	l.Element = l.Element.build(scope)
	return l
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
			if tuple.elements[i] != nil && !tuple.elements[i].Extends(t.elements[i]) {
				return false
			}
		}
		return true
	default:
		if len(tuple.elements) == 1 {
			return tuple.elements[0].Extends(t)
		}
		return false
	}
}
func (t Tuple) build(scope *Scope) ExpressionType {
	for i, el := range t.elements {
		t.elements[i] = el.build(scope)
	}
	return t
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
func (r Range) build(scope *Scope) ExpressionType {
	r.operands = r.operands.build(scope)
	return r
}

type Function struct {
	TypeParams []string
	Params     Tuple
	Returned   ExpressionType
}

func (f Function) Kind() ExpressionTypeKind      { return FUNCTION }
func (f Function) Match(t ExpressionType) bool   { /* FIXME: */ return false }
func (f Function) Extends(t ExpressionType) bool { /* FIXME: */ return false }

// FIXME: generics
func (f Function) build(scope *Scope) ExpressionType { return f }

type Object struct {
	Members map[string]ExpressionType
}

func (o Object) Kind() ExpressionTypeKind    { return STRUCT }
func (o Object) Match(t ExpressionType) bool { return false }
func (o Object) Extends(t ExpressionType) bool {
	structB, ok := t.(Object)
	if !ok {
		return false
	}
	for member, typeA := range o.Members {
		typeB, ok := structB.Members[member]
		if !ok {
			return false
		}
		if !typeA.Extends(typeB) {
			return false
		}
	}
	for member := range structB.Members {
		if _, ok := o.Members[member]; !ok {
			return false
		}
	}
	return true
}

// FIXME:
func (o Object) build(scope *Scope) ExpressionType { return o }

type Deferred struct{}

func (d Deferred) Kind() ExpressionTypeKind          { return UNKNOWN }
func (d Deferred) Match(t ExpressionType) bool       { return true }
func (d Deferred) Extends(t ExpressionType) bool     { return true }
func (d Deferred) build(scope *Scope) ExpressionType { return d }

type Generic struct {
	Name string
	Ref  *Variable
}

func (g Generic) Kind() ExpressionTypeKind          { return UNKNOWN }
func (g Generic) Match(t ExpressionType) bool       { return true }
func (g Generic) Extends(t ExpressionType) bool     { return true }
func (g Generic) build(scope *Scope) ExpressionType { return g }

func ReadTypeExpression(expr parser.Node) ExpressionType {
	switch expr := expr.(type) {
	case Literal:
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
