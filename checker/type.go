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
	SUM

	FUNCTION
	CONSTRUCTOR
	TRAIT
)

type ExpressionType interface {
	Kind() ExpressionTypeKind
	Match(ExpressionType) bool
	Extends(ExpressionType) bool
	build(*Scope, ExpressionType) (ExpressionType, bool)
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
func (t Type) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	var value ExpressionType
	if c, ok := compared.(Type); ok {
		value = c.Value
	}
	v, ok := t.Value.build(scope, value)
	return Type{v}, ok
}

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
func (p Primitive) build(scope *Scope, c ExpressionType) (ExpressionType, bool) { return p, true }

type TypeAlias struct {
	Name    string
	Params  []Generic
	Ref     ExpressionType
	Methods map[string]ExpressionType
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
		if param.Value != nil && !param.Value.Match(alias.Params[i]) {
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
	if alias.Name != ta.Name {
		return false
	}
	for i, param := range ta.Params {
		if param.Value != nil && !param.Value.Extends(alias.Params[i]) {
			return false
		}
	}
	return true
}

func (ta TypeAlias) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	s := NewScope()
	s.outer = scope
	for _, param := range ta.Params {
		s.Add(param.Name, tokenizer.Loc{}, param)
	}
	var ref ExpressionType
	if c, ok := compared.(TypeAlias); ok {
		ref = c.Ref
	}
	ref, ok := ta.Ref.build(s, ref)
	ta.Ref = ref
	return ta, ok
}

func (ta *TypeAlias) registerMethod(name string, signature ExpressionType) {
	if ta.Methods == nil {
		ta.Methods = map[string]ExpressionType{}
	}
	ta.Methods[name] = signature
}
func (ta TypeAlias) implements(trait Trait) bool {
	for name, signature := range trait.Members {
		method, ok := ta.Methods[name]
		if !ok || !signature.Extends(method) {
			return false
		}
	}
	return true
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
func (l List) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	var element ExpressionType
	if c, ok := compared.(List); ok {
		element = c.Element
	}
	var ok bool
	l.Element, ok = l.Element.build(scope, element)
	return l, ok
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

// FIXME: indexes
func (t Tuple) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	ok := true
	c, k := compared.(Tuple)
	if compared == nil || !k {
		for i, el := range t.elements {
			t.elements[i], k = el.build(scope, nil)
			ok = ok && k
		}
		return t, ok
	}
	for i, el := range t.elements {
		t.elements[i], k = el.build(scope, c.elements[i])
		ok = ok && k
	}
	return t, ok
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
func (r Range) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	var operands ExpressionType
	if c, ok := compared.(Range); ok {
		operands = c.operands
	}
	operands, ok := r.operands.build(scope, operands)
	return Range{operands}, ok
}

type Function struct {
	TypeParams []Generic
	Params     Tuple
	Returned   ExpressionType
}

func (f Function) Kind() ExpressionTypeKind    { return FUNCTION }
func (f Function) Match(t ExpressionType) bool { /* FIXME: */ return false }
func (f Function) Extends(t ExpressionType) bool {
	function, ok := t.(Function)
	if !ok {
		return false
	}
	if len(f.Params.elements) != len(function.Params.elements) {
		return false
	}
	for i, param := range f.Params.elements {
		if !param.Extends(function.Params.elements[i]) {
			return false
		}
	}
	if (f.Returned == nil) != (function.Returned == nil) {
		return false
	}
	return f.Returned == nil || f.Returned.Extends(function.Returned)
}

func (f Function) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	ok := true
	s := NewScope()
	s.outer = scope
	for _, param := range f.TypeParams {
		s.Add(param.Name, tokenizer.Loc{}, param)
	}
	c, k := compared.(Function)
	f.Params = Tuple{make([]ExpressionType, len(f.Params.elements))}
	for i, param := range f.Params.elements {
		var el ExpressionType
		if len(c.Params.elements) > i {
			el = c.Params.elements[i]
		}
		p, k := param.build(s, el)
		ok = ok && k
		f.Params.elements[i] = p
	}
	var r ExpressionType
	if k {
		r = c.Returned
	}
	f.Returned, k = f.Returned.build(s, r)
	ok = ok && k
	return f, ok
}

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

func (o Object) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	ok := true
	for name, member := range o.Members {
		var k bool
		// FIXME: is compared an object?
		o.Members[name], k = member.build(scope, compared)
		ok = ok && k
	}
	return o, ok
}

type Sum struct {
	Members map[string]ExpressionType
}

func (s Sum) Kind() ExpressionTypeKind    { return SUM }
func (s Sum) Match(t ExpressionType) bool { return false }
func (s Sum) Extends(t ExpressionType) bool {
	// return true if exactly one member extends received type
	found := false
	for _, member := range s.Members {
		if !member.Extends(t) {
			continue
		}
		if found {
			return false
		}
		found = true
	}
	return found
}
func (s Sum) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	ok := true
	for name, member := range s.Members {
		var k bool
		// FIXME: is compared a sum type? should it work like this?
		s.Members[name], k = member.build(scope, compared)
		ok = ok && k
	}
	return s, ok
}

type Trait struct {
	Self    Generic
	Members map[string]ExpressionType
}

func (t Trait) Kind() ExpressionTypeKind     { return TRAIT }
func (t Trait) Match(et ExpressionType) bool { return false }
func (t Trait) Extends(et ExpressionType) bool {
	var receivedMethods map[string]ExpressionType
	switch et := et.(type) {
	case Trait:
		receivedMethods = et.Members
	case TypeAlias:
		receivedMethods = et.Methods
	default:
		return false
	}
	for name, signature := range t.Members {
		method, ok := receivedMethods[name]
		if !ok || !signature.Extends(method) {
			return false
		}
	}
	return true
}
func (t Trait) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	// FIXME:
	return t, true
}

type Generic struct {
	Name        string
	Constraints ExpressionType
	Value       ExpressionType
}

func (g Generic) Kind() ExpressionTypeKind {
	if g.Constraints == nil {
		return UNKNOWN
	}
	return g.Constraints.Kind()
}
func (g Generic) Match(t ExpressionType) bool {
	if g.Constraints == nil {
		return true
	}
	return g.Constraints.Match(t)
}
func (g Generic) Extends(t ExpressionType) bool {
	if g.Constraints == nil {
		return true
	}
	return g.Constraints.Extends(t)
}
func (g Generic) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	variable, ok := scope.Find(g.Name)
	if !ok {
		return Primitive{UNKNOWN}, false
	}
	variable.reads = append(variable.reads, tokenizer.Loc{})
	ok = isGenericType(variable.typing)
	if !ok {
		return variable.typing, true
	}
	t := variable.typing.(Type)
	generic := t.Value.(Generic)
	if generic.Value == nil {
		generic.Value = compared
	}
	t.Value = generic
	variable.typing = t
	return generic.Value, generic.Value != nil
}
func isGenericType(typing ExpressionType) bool {
	t, ok := typing.(Type)
	if !ok {
		return false
	}
	_, ok = t.Value.(Generic)
	return ok
}

type Constructor struct {
	From ExpressionType
	To   ExpressionType
}

func (c Constructor) Kind() ExpressionTypeKind      { return CONSTRUCTOR }
func (c Constructor) Match(t ExpressionType) bool   { return false }
func (c Constructor) Extends(t ExpressionType) bool { return false }
func (c Constructor) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	done := true
	co, ok := compared.(Constructor)
	if ok {
		c.From, ok = c.From.build(scope, co.From)
		done = done && ok
		c.To, ok = c.To.build(scope, co.To)
		done = done && ok
	} else {
		c.From, ok = c.From.build(scope, nil)
		done = done && ok
		c.To, ok = c.To.build(scope, nil)
		done = done && ok
	}
	return c, done
}

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
