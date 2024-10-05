package checker

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type Variable struct {
	declaredAt tokenizer.Loc
	typing     ExpressionType
	writes     []tokenizer.Loc
	reads      []tokenizer.Loc
}

func (v *Variable) readAt(loc tokenizer.Loc)  { v.reads = append(v.reads, loc) }
func (v *Variable) writeAt(loc tokenizer.Loc) { v.writes = append(v.writes, loc) }

type Method struct {
	self       ExpressionType
	signature  Function
	declaredAt tokenizer.Loc
	reads      []tokenizer.Loc
}

type ScopeKind int8

const (
	ProgramScope ScopeKind = iota
	BlockScope
	FunctionScope
	LoopScope
)

type Scope struct {
	variables map[string]*Variable
	methods   map[string][]Method
	kind      ScopeKind
	outer     *Scope
	shadow    bool
}

func NewScope(kind ScopeKind) *Scope {
	return &Scope{
		kind:      kind,
		variables: map[string]*Variable{},
		methods:   map[string][]Method{},
	}
}

func NewShadowScope() *Scope {
	return &Scope{
		variables: map[string]*Variable{},
		methods:   map[string][]Method{},
		shadow:    true,
	}
}

func (s Scope) Find(name string) (*Variable, bool) {
	variable, ok := s.variables[name]
	if ok {
		return variable, true
	}
	if s.outer != nil {
		return s.outer.Find(name)
	}
	return nil, false
}

func (s Scope) FindMethod(name string, typing ExpressionType) (*Method, bool) {
	for _, method := range s.methods[name] {
		found := method.self
		if f, ok := found.(Type); ok {
			found = f.Value
		}
		if found.Match(typing) {
			return &method, true
		}
	}
	if s.outer != nil {
		return s.outer.FindMethod(name, typing)
	}
	return nil, false
}

func (s Scope) Has(name string) bool {
	_, ok := s.variables[name]
	if ok {
		return true
	}
	if s.outer != nil && s.outer.shadow {
		return s.outer.Has(name)
	}
	return false
}

func (s *Scope) Add(name string, declaredAt tokenizer.Loc, typing ExpressionType) {
	if name == "" || name == "_" {
		return
	}
	s.variables[name] = &Variable{
		declaredAt: declaredAt,
		typing:     typing,
	}
}

func (s *Scope) AddMethod(name string, declaredAt tokenizer.Loc, self ExpressionType, signature Function) {
	s.methods[name] = append(s.methods[name], Method{self, signature, declaredAt, []tokenizer.Loc{}})
}

func (s *Scope) WriteAt(name string, loc tokenizer.Loc) {
	variable, ok := s.Find(name)
	// TODO: panic on error
	if ok {
		variable.writeAt(loc)
	}
}

func (s *Scope) ReadAt(name string, loc tokenizer.Loc) {
	variable, ok := s.Find(name)
	// TODO: panic on error
	if ok {
		variable.readAt(loc)
	}
}

func (s Scope) in(kind ScopeKind) bool {
	if s.kind == kind {
		return true
	}
	if s.outer == nil {
		return false
	}
	return s.outer.in(kind)
}
