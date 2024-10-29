package parser

type Variable struct {
	declaredAt Loc
	typing     ExpressionType
	writes     []Loc
	reads      []Loc
}

func (v *Variable) readAt(loc Loc)  { v.reads = append(v.reads, loc) }
func (v *Variable) writeAt(loc Loc) { v.writes = append(v.writes, loc) }

type ScopeKind int8

const (
	ProgramScope ScopeKind = iota
	BlockScope
	FunctionScope
	LoopScope
)

type Scope struct {
	variables map[string]*Variable
	kind      ScopeKind
	outer     *Scope
}

func NewScope(kind ScopeKind) *Scope {
	return &Scope{
		kind:      kind,
		variables: map[string]*Variable{},
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
		typing:     typing,
	}
}

func (s *Scope) AddMethod(name string, self TypeAlias, signature Function) {
	t, ok := s.Find(self.Name)
	if !ok {
		return
	}
	self.registerMethod(name, signature)
	t.typing = Type{self}
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
	alias := TypeAlias{
		Name:   "Option",
		Params: []Generic{{Name: "Type", Value: t}},
		Ref: Sum{map[string]Function{
			"Some": {
				Params: &Tuple{[]ExpressionType{Generic{Name: "Type", Value: t}}},
			},
			"None": {},
		}},
	}
	some := alias.Ref.(Sum).Members["Some"]
	some.Returned = alias
	alias.Ref.(Sum).Members["Some"] = some

	none := alias.Ref.(Sum).Members["None"]
	none.Returned = alias
	alias.Ref.(Sum).Members["None"] = none
	return alias
}

// extracts the Some type from an Option
func getSomeType(t Sum) ExpressionType {
	some := t.Members["Some"]
	return some.Params.elements[0].(Generic).Value
}

// The vanilla option type.
// It represents the presence or absence of some value.
var optionType = makeOptionType(nil)

// utility to create result types
func makeResultType(ok ExpressionType, err ExpressionType) TypeAlias {
	alias := TypeAlias{
		Name: "Result",
		Params: []Generic{
			{Name: "Ok", Value: ok},
			{Name: "Err", Value: err},
		},
		Ref: Sum{map[string]Function{
			"Ok": {
				Params: &Tuple{[]ExpressionType{Generic{Name: "Ok", Value: ok}}},
			},
			"Err": {
				Params: &Tuple{[]ExpressionType{Generic{Name: "Err", Value: err}}},
			},
		}},
	}
	okConst := alias.Ref.(Sum).Members["Ok"]
	okConst.Returned = alias
	alias.Ref.(Sum).Members["Ok"] = okConst

	errConst := alias.Ref.(Sum).Members["Err"]
	errConst.Returned = alias
	alias.Ref.(Sum).Members["Err"] = errConst
	return alias
}

// utility to create map types
func makeMapType(key ExpressionType, value ExpressionType) TypeAlias {
	alias := TypeAlias{
		Name: "Map",
		Params: []Generic{
			{Name: "Key", Value: key},
			{Name: "Value", Value: value},
		},
		Ref: Map{
			Generic{Name: "Key", Value: key},
			Generic{Name: "Value", Value: value},
		},
	}
	return alias
}

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
		"Map": {
			typing: Type{makeMapType(nil, nil)},
		},
		"Option": {
			typing: Type{optionType},
		},
		"Result": {
			typing: Type{makeResultType(nil, nil)},
		},
	},
}
