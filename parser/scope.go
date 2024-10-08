package parser

type Variable struct {
	declaredAt Loc
	typing     ExpressionType
	writes     []Loc
	reads      []Loc
}

func (v *Variable) readAt(loc Loc)  { v.reads = append(v.reads, loc) }
func (v *Variable) writeAt(loc Loc) { v.writes = append(v.writes, loc) }

type Method struct {
	self       ExpressionType
	signature  Function
	declaredAt Loc
	reads      []Loc
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

func (s *Scope) Add(name string, declaredAt Loc, typing ExpressionType) {
	if name == "" || name == "_" {
		return
	}
	s.variables[name] = &Variable{
		declaredAt: declaredAt,
		typing:     typing,
	}
}

func (s *Scope) AddMethod(name string, declaredAt Loc, self ExpressionType, signature Function) {
	s.methods[name] = append(s.methods[name], Method{self, signature, declaredAt, []Loc{}})
}

func (s *Scope) WriteAt(name string, loc Loc) {
	variable, ok := s.Find(name)
	// TODO: panic on error
	if ok {
		variable.writeAt(loc)
	}
}

func (s *Scope) ReadAt(name string, loc Loc) {
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

// utility to create option types with different Some types
func makeOptionType(t ExpressionType) TypeAlias {
	return TypeAlias{
		Name:   "Option",
		Params: []Generic{{Name: "Type", Value: t}},
		Ref: Sum{map[string]ExpressionType{
			"Some": Generic{Name: "Type", Value: t},
			"None": nil,
		}},
	}
}

// The vanilla option type.
// It represents the presence or absence of some value.
var optionType = makeOptionType(nil)

// The scope containing the standard library
var std = Scope{
	variables: map[string]*Variable{
		"List": {
			typing: Type{TypeAlias{
				Name:   "List",
				Params: []Generic{{Name: "Type"}},
				Ref:    List{Generic{Name: "Type"}},
			}},
		},
		"Option": {
			typing: Type{optionType},
		},
		"Result": {
			typing: Type{TypeAlias{
				Name:   "Result",
				Params: []Generic{{Name: "Ok"}, {Name: "Err"}},
				Ref: Sum{map[string]ExpressionType{
					"Ok":  Type{Generic{Name: "Ok"}},
					"Err": Type{Generic{Name: "Err"}},
				}},
			}},
		},
	},
}
