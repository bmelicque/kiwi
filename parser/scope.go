package parser

type Variable struct {
	declaredAt   Loc
	Typing       ExpressionType
	scope        *Scope
	writes       []Node
	reads        []Loc
	hasDirectRef bool
}

func (v *Variable) readAt(l Loc)   { v.reads = append(v.reads, l) }
func (v *Variable) writeAt(n Node) { v.writes = append(v.writes, n) }

func (v *Variable) Writes() []Node     { return v.writes }
func (v *Variable) HasDirectRef() bool { return v.hasDirectRef }

type ScopeKind uint8

const (
	ProgramScope ScopeKind = iota
	BlockScope
	FunctionScope
	LoopScope
)

type Scope struct {
	id        int
	variables map[string]*Variable
	kind      ScopeKind
	outer     *Scope
}

var lastScopeId int

func NewScope(kind ScopeKind) *Scope {
	lastScopeId++
	return &Scope{
		id:        lastScopeId,
		kind:      kind,
		variables: map[string]*Variable{},
	}
}

func (s Scope) GetId() int { return s.id }

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

func (s Scope) FindLocal(name string) *Variable {
	return s.variables[name]
}

func (s Scope) Has(name string) bool {
	_, ok := s.variables[name]
	if ok {
		return true
	}
	if s.outer != nil {
		return s.outer.Has(name)
	}
	return false
}

func (s *Scope) Add(name string, declaredAt Loc, typing ExpressionType) {
	if name == "" || name == "_" {
		return
	}
	s.variables[name] = &Variable{
		declaredAt: declaredAt,
		Typing:     typing,
		scope:      s,
	}
}

func (s *Scope) AddMethod(name string, self TypeAlias, signature Function) {
	t, ok := s.Find(self.Name)
	if !ok {
		return
	}
	self.registerMethod(name, signature)
	t.Typing = Type{self}
}

func (s Scope) HasReferencedVars() bool {
	for _, v := range s.variables {
		if v.hasDirectRef {
			return true
		}
	}
	return false
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

func (s Scope) toModule() Module {
	o := newObject()
	for name, v := range s.variables {
		if name[0] != '_' {
			o.addMember(name, v.Typing)
		}
	}
	return Module{o}
}

// utility to create option types with different Some types
func makeOptionType(t ExpressionType) TypeAlias {
	return TypeAlias{
		Name:   "?",
		Params: []Generic{{Name: "Type", Value: t}},
		Ref: Sum{map[string]Tuple{
			"Some": {[]ExpressionType{Generic{Name: "Type", Value: t}}},
			"None": newTuple(),
		}},
	}
}

// extracts the Some type from an Option
func getSomeType(t Sum) ExpressionType {
	some := t.Members["Some"]
	return some.Elements[0].(Generic).Value
}

// The vanilla option type.
// It represents the presence or absence of some value.
var optionType = makeOptionType(nil)

// utility to create result types
func makeResultType(ok ExpressionType, err ExpressionType) TypeAlias {
	return TypeAlias{
		Name: "!",
		Params: []Generic{
			{Name: "Ok", Value: ok},
			{Name: "Err", Value: err},
		},
		Ref: Sum{map[string]Tuple{
			"Ok":  {[]ExpressionType{Generic{Name: "Ok", Value: ok}}},
			"Err": {[]ExpressionType{Generic{Name: "Err", Value: err}}},
		}},
	}
}

var mapMethods = map[string]ExpressionType{
	"has": Function{
		Params:   &Tuple{[]ExpressionType{Generic{Name: "Key"}}},
		Returned: Boolean{},
	},
	"get": Function{
		Params:   &Tuple{[]ExpressionType{Generic{Name: "Key"}}},
		Returned: makeOptionType(Generic{Name: "Value"}),
	},
	"set": Function{
		Params:   &Tuple{[]ExpressionType{Generic{Name: "Key"}, Generic{Name: "Value"}}},
		Returned: Void{},
	},
}

// utility to create map types
func makeMapType(key ExpressionType, value ExpressionType) TypeAlias {
	alias := TypeAlias{
		Name: "#",
		Params: []Generic{
			{Name: "Key", Value: key},
			{Name: "Value", Value: value},
		},
		Ref: Map{
			Generic{Name: "Key", Value: key},
			Generic{Name: "Value", Value: value},
		},
		Methods: mapMethods,
	}
	return alias
}

// utility to create map types
func makePromise(t ExpressionType) TypeAlias {
	return TypeAlias{
		Name:   "...",
		Params: []Generic{{Name: "Type", Value: t}},
		Ref: Object{
			Members: []ObjectMember{
				{"value", Generic{Name: "Type", Value: t}},
			},
			Defaults: []ObjectMember{},
		},
	}
}

// The scope containing the standard library
var std = Scope{
	variables: map[string]*Variable{
		"Error": {
			Typing: Type{Trait{Members: map[string]ExpressionType{"error": newGetter(String{})}}},
		},
		"#": {
			Typing: Type{makeMapType(nil, nil)},
		},
		"...": {
			Typing: Type{makePromise(nil)},
		},
		"?": {
			Typing: Type{optionType},
		},
		"!": {
			Typing: Type{makeResultType(nil, nil)},
		},
	},
}
