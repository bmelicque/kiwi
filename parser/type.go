package parser

type ExpressionType interface {
	Extends(ExpressionType) bool
	Text() string
	build(*Scope, ExpressionType) (ExpressionType, bool)
}

func Match(a ExpressionType, b ExpressionType) bool {
	if a == nil || b == nil {
		return true
	}
	if _, ok := a.(Unknown); ok {
		return true
	}
	if _, ok := b.(Unknown); ok {
		return true
	}
	return a.Extends(b) && b.Extends(a)
}

type Type struct {
	Value ExpressionType
}

func (t Type) Extends(testType ExpressionType) bool {
	got, ok := testType.(Type)
	return ok && t.Value.Extends(got.Value)
}
func (t Type) Text() string { return "(" + t.Value.Text() + ")" }
func (t Type) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	var value ExpressionType
	if c, ok := compared.(Type); ok {
		value = c.Value
	}
	v, ok := t.Value.build(scope, value)
	return Type{v}, ok
}

type Unknown struct{}

func (u Unknown) Extends(t ExpressionType) bool { return true }
func (u Unknown) Text() string                  { return "unknown" }
func (u Unknown) build(scope *Scope, c ExpressionType) (ExpressionType, bool) {
	return u, true
}

type Nil struct{}

func (n Nil) Extends(t ExpressionType) bool {
	_, ok := t.(Number)
	return ok
}
func (n Nil) Text() string { return "nil" }
func (n Nil) build(scope *Scope, c ExpressionType) (ExpressionType, bool) {
	return n, true
}

type Number struct{}

func (n Number) Extends(t ExpressionType) bool {
	_, ok := t.(Number)
	return ok
}
func (n Number) Text() string { return "number" }
func (n Number) build(scope *Scope, c ExpressionType) (ExpressionType, bool) {
	return n, true
}

type Boolean struct{}

func (b Boolean) Extends(t ExpressionType) bool {
	_, ok := t.(Boolean)
	return ok
}
func (b Boolean) Text() string { return "boolean" }
func (b Boolean) build(scope *Scope, c ExpressionType) (ExpressionType, bool) {
	return b, true
}

type String struct{}

func (s String) Extends(t ExpressionType) bool {
	_, ok := t.(String)
	return ok
}
func (s String) Text() string { return "string" }
func (s String) build(scope *Scope, c ExpressionType) (ExpressionType, bool) {
	return s, true
}

type TypeAlias struct {
	Name    string
	Params  []Generic
	Ref     ExpressionType
	Methods map[string]ExpressionType
	from    string // path to origin file
}

func (ta TypeAlias) Extends(t ExpressionType) bool {
	alias, ok := t.(TypeAlias)
	if !ok {
		return ta.Ref.Extends(t)
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
func (ta TypeAlias) Text() string {
	params := []Generic{}
	for _, param := range ta.Params {
		if param.Value == nil {
			break
		}
		params = append(params, param)
	}
	if len(params) == 0 {
		return ta.Name
	}
	var s string
	if ta.Name == "..." {
		s = "async"
	} else {
		s = ta.Name
	}
	s += "["
	max := len(params) - 1
	for _, param := range params[:max] {
		s += param.Value.Text()
		s += ", "
	}
	s += params[max].Value.Text()
	return s + "]"
}
func (ta TypeAlias) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	s := NewScope(ProgramScope)
	s.outer = scope
	for _, param := range ta.Params {
		s.Add(param.Name, Loc{}, param)
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

type Ref struct {
	To ExpressionType
}

func (r Ref) Extends(t ExpressionType) bool {
	ref, ok := t.(Ref)
	if !ok {
		return false
	}
	return r.To.Extends(ref.To)
}
func (r Ref) Text() string { return "&" + r.To.Text() }
func (r Ref) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	ref, ok := compared.(Ref)
	if !ok {
		return r, false
	}
	r.To, ok = r.To.build(scope, ref.To)
	return r, ok
}
func deref(t ExpressionType) ExpressionType {
	if ref, ok := t.(Ref); ok {
		return ref.To
	}
	return t
}

type List struct {
	Element ExpressionType
}

func (l List) Extends(t ExpressionType) bool {
	if list, ok := t.(List); ok {
		return l.Element.Extends(list.Element)
	}
	return false
}
func (l List) Text() string { return "[]" + l.Element.Text() }
func (l List) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	var element ExpressionType
	if c, ok := compared.(List); ok {
		element = c.Element
	}
	var ok bool
	l.Element, ok = l.Element.build(scope, element)
	return l, ok
}

type Map struct {
	Key   ExpressionType
	Value ExpressionType
}

func (m Map) Extends(received ExpressionType) bool {
	t, ok := received.(Map)
	if !ok {
		return false
	}
	return m.Key.Extends(t.Key) && m.Value.Extends(t.Value)
}
func (m Map) Text() string { return "Map[" + m.Key.Text() + ", " + m.Value.Text() + "]" }
func (m Map) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	c, ok := compared.(Map)
	if !ok {
		key, kk := m.Key.build(scope, nil)
		value, vk := m.Value.build(scope, nil)
		return Map{key, value}, kk && vk
	}
	key, kk := m.Key.build(scope, c.Key)
	value, vk := m.Value.build(scope, c.Value)
	return Map{key, value}, kk && vk
}

type Tuple struct {
	Elements []ExpressionType
}

func (tuple Tuple) Extends(t ExpressionType) bool {
	switch t := t.(type) {
	case Tuple:
		if len(t.Elements) != len(tuple.Elements) {
			return false
		}
		for i := 0; i < len(t.Elements); i += 1 {
			if tuple.Elements[i] != nil && !tuple.Elements[i].Extends(t.Elements[i]) {
				return false
			}
		}
		return true
	default:
		if len(tuple.Elements) == 1 {
			return tuple.Elements[0].Extends(t)
		}
		return false
	}
}
func (t Tuple) Text() string {
	s := "("
	max := len(t.Elements) - 1
	for _, el := range t.Elements[:max] {
		s += el.Text() + ", "
	}
	return s + t.Elements[max].Text() + ")"
}

// FIXME: indexes
func (t Tuple) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	ok := true
	c, k := compared.(Tuple)
	if compared == nil || !k {
		for i, el := range t.Elements {
			t.Elements[i], k = el.build(scope, nil)
			ok = ok && k
		}
		return t, ok
	}
	for i, el := range t.Elements {
		t.Elements[i], k = el.build(scope, c.Elements[i])
		ok = ok && k
	}
	return t, ok
}

type Range struct {
	operands ExpressionType
}

func (r Range) Extends(t ExpressionType) bool {
	if received, ok := t.(Range); ok {
		return r.operands.Extends(received.operands)
	}
	return false
}
func (r Range) Text() string { return ".." + r.operands.Text() }
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
	Params     *Tuple
	Returned   ExpressionType
	Async      bool // true if the function can be called with 'async'
}

func (f Function) arity() int {
	if f.Params == nil {
		return 0
	}
	return len(f.Params.Elements)
}

func (f Function) Extends(t ExpressionType) bool {
	function, ok := t.(Function)
	if !ok {
		return false
	}
	if f.arity() != function.arity() {
		return false
	}
	if f.arity() == 0 {
		return true
	}
	for i, param := range f.Params.Elements {
		if !param.Extends(function.Params.Elements[i]) {
			return false
		}
	}
	if (f.Returned == nil) != (function.Returned == nil) {
		return false
	}
	return f.Returned == nil || f.Returned.Extends(function.Returned)
}
func (f Function) Text() string { return f.Params.Text() + " -> " + f.Returned.Text() }
func (f Function) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	ok := true
	s := NewScope(ProgramScope)
	s.outer = scope
	for _, param := range f.TypeParams {
		s.Add(param.Name, Loc{}, param)
	}
	c, k := compared.(Function)
	f.Params = &Tuple{make([]ExpressionType, len(f.Params.Elements))}
	for i, param := range f.Params.Elements {
		var el ExpressionType
		if len(c.Params.Elements) > i {
			el = c.Params.Elements[i]
		}
		p, k := param.build(s, el)
		ok = ok && k
		f.Params.Elements[i] = p
	}
	var r ExpressionType
	if k {
		r = c.Returned
	}
	f.Returned, k = f.Returned.build(s, r)
	ok = ok && k
	return f, ok
}

type ObjectMember struct {
	Name string
	Type ExpressionType
}

type Object struct {
	Members  []ObjectMember
	Defaults []ObjectMember
}

func newObject() Object {
	return Object{
		Members:  []ObjectMember{},
		Defaults: []ObjectMember{},
	}
}

func (o Object) Extends(t ExpressionType) bool { return false }
func (o Object) Text() string {
	s := "{"
	for _, member := range o.Members {
		s += member.Name + ": " + member.Type.Text() + ", "
	}
	return s + "}"
}
func (o Object) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	ok := true
	for i, member := range o.Members {
		var k bool
		// FIXME: is compared an object?
		built, k := member.Type.build(scope, compared)
		o.Members[i] = ObjectMember{member.Name, built}
		ok = ok && k
	}
	// FIXME: Defaults
	return o, ok
}

func (o Object) get(name string) (ExpressionType, bool) {
	for _, member := range o.Members {
		if member.Name == name {
			return member.Type, true
		}
	}
	for _, member := range o.Defaults {
		if member.Name == name {
			return member.Type, true
		}
	}
	return nil, false
}

func (o *Object) addMember(name string, t ExpressionType) {
	o.Members = append(o.Members, ObjectMember{name, t})
}
func (o *Object) addDefault(name string, t ExpressionType) {
	o.Defaults = append(o.Defaults, ObjectMember{name, t})
}

type Module struct {
	Object
}

type Sum struct {
	Members map[string]Function
}

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
func (s Sum) Text() string {
	str := "("
	for name, member := range s.Members {
		str += "| " + name + member.Params.Text() + " "
	}
	return str + ")"
}
func (s Sum) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	ok := true
	for name, member := range s.Members {
		var k bool
		// FIXME: is compared a sum type? should it work like this?
		m, k := member.build(scope, compared)
		s.Members[name] = m.(Function)
		ok = ok && k
	}
	return s, ok
}
func (s Sum) getMember(name string) ExpressionType {
	member, ok := s.Members[name]
	if !ok {
		return Unknown{}
	}
	if len(member.Params.Elements) == 1 {
		ret, _ := member.Params.Elements[0].build(nil, nil)
		return ret
	}
	tuple := Tuple{make([]ExpressionType, len(member.Params.Elements))}
	for i := range member.Params.Elements {
		tuple.Elements[i], _ = member.Params.Elements[i].build(nil, nil)
	}
	return tuple
}

type Trait struct {
	Self    Generic
	Members map[string]ExpressionType
}

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
func (t Trait) Text() string {
	s := "("
	for name, member := range t.Members {
		s += name + ": " + member.Text() + ", "
	}
	return s + ")"
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

func (g Generic) Extends(t ExpressionType) bool {
	if g.Constraints == nil {
		return true
	}
	return g.Constraints.Extends(t)
}
func (g Generic) Text() string { return g.Name }
func (g Generic) build(scope *Scope, compared ExpressionType) (ExpressionType, bool) {
	if g.Value != nil {
		return g.Value, true
	}
	if scope == nil {
		return Unknown{}, false
	}
	variable, ok := scope.Find(g.Name)
	if !ok {
		return Unknown{}, false
	}
	variable.readAt(Loc{})
	ok = isGenericType(variable.Typing)
	if !ok {
		return variable.Typing, true
	}
	t := variable.Typing.(Type)
	generic := t.Value.(Generic)
	if generic.Value == nil {
		generic.Value = compared
	}
	t.Value = generic
	variable.Typing = t
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
