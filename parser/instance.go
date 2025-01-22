package parser

import (
	"fmt"
)

type InstanceExpression struct {
	Typing Expression
	Args   *BracedExpression
	typing ExpressionType
}

func (i *InstanceExpression) Type() ExpressionType { return i.typing }

func (i *InstanceExpression) getChildren() []Node {
	children := []Node{i.Typing}
	if i.Args != nil {
		children = append(children, i.Args)
	}
	return children
}

func (i *InstanceExpression) typeCheck(p *Parser) {
	if checkInferredInstances(p, i) {
		return
	}
	if !checkInstanceConstructor(p, i) {
		i.Args.typeCheck(p)
		return
	}
	constructor, isRef := getConstructorType(i)
	switch constructor.(type) {
	case TypeAlias:
		checkObjectInstanciation(p, i)
	case List:
		typeCheckNamedListInstanciation(p, i)
	default:
		p.error(i.Typing, NotInstanceable, constructor)
	}
	if isRef {
		i.typing = Ref{i.typing}
	}
}
func checkInferredInstances(p *Parser, i *InstanceExpression) bool {
	switch {
	case isInferredList(i):
		checkInferredListInstance(p, i)
		return true
	case isInferredOption(i):
		checkInferredOptionInstance(p, i)
		return true
	case isInferredMap(i):
		checkInferredMapInstance(p, i)
		return true
	default:
		return false
	}
}
func isInferredList(i *InstanceExpression) bool {
	l, ok := i.Typing.(*ListTypeExpression)
	return ok && l.Expr == nil
}
func isInferredOption(i *InstanceExpression) bool {
	u, ok := i.Typing.(*UnaryExpression)
	return ok && u.Operator.Kind() == QuestionMark && u.Operand == nil
}
func isInferredMap(i *InstanceExpression) bool {
	b, ok := i.Typing.(*BinaryExpression)
	return ok && b.Operator.Kind() == Hash && b.Left == nil && b.Right == nil
}

func checkInstanceConstructor(p *Parser, i *InstanceExpression) bool {
	i.Typing.typeCheck(p)
	_, ok := i.Typing.Type().(Type)
	if !ok {
		p.error(i.Typing, TypeExpected)
	}
	return ok
}

func getConstructorType(i *InstanceExpression) (ExpressionType, bool) {
	t := i.Typing.Type().(Type)
	constructor := t.Value
	ref, isRef := constructor.(Ref)
	if isRef {
		constructor = ref.To
	}
	return constructor, isRef
}

// Parse a struct instanciation, like 'Object{key: value}'
func checkObjectInstanciation(p *Parser, i *InstanceExpression) {
	// next line should be ensured by calling function
	t := i.Typing.Type().(Type).Value
	var alias TypeAlias
	if ref, ok := t.(Ref); ok {
		alias = ref.To.(TypeAlias)
	} else {
		alias = t.(TypeAlias)
	}
	if alias.Name == "?" {
		checkOptionInstanciation(p, i, alias)
		return
	}
	if alias.Name == "#" {
		typeCheckMapInstanciation(p, i, alias)
		return
	}
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	typeCheckTypeArgs(p, nil, alias.Params)

	object, ok := alias.Ref.(Object)
	if !ok {
		p.error(i.Typing, ObjectTypeExpected)
		i.typing = Invalid{}
		return
	}

	args := i.Args.Expr.(*TupleExpression).Elements
	formatStructEntries(p, args)
	for _, arg := range args {
		entry := arg.(*Entry)
		var name string
		if entry.Key != nil {
			name = entry.Key.(*Identifier).Text()
		}
		expected, ok := object.GetOwned(name)
		if ok && entry.Value != nil && !expected.Extends(entry.Value.Type()) {
			p.error(arg, CannotAssignType, expected, entry.Value.Type())
		}
	}

	reportExcessMembers(p, object, args)
	reportMissingMembers(p, object, i.Args)

	i.typing = alias
}
func formatStructEntries(p *Parser, received []Expression) {
	for i := range received {
		received[i] = getFormattedStructEntry(p, received[i])
	}
}

// Format and report non-entry expression as a valid *Entry
func getFormattedStructEntry(p *Parser, received Expression) *Entry {
	entry, ok := received.(*Entry)
	if !ok {
		if param, ok := received.(*Param); ok {
			p.error(param, FieldExpected)
			return &Entry{Key: param.Identifier, Value: param.Complement}
		} else {
			p.error(received, FieldExpected)
			return &Entry{Value: received}
		}
	}
	if _, ok := entry.Key.(*Identifier); !ok {
		p.error(entry.Key, IdentifierExpected)
		entry.Key = &Identifier{}
	}
	return entry
}
func reportExcessMembers(p *Parser, expected Object, received []Expression) {
	for _, arg := range received {
		namedArg, ok := arg.(*Entry)
		if !ok || namedArg.Key == nil {
			continue
		}
		name := namedArg.Key.(*Identifier).Text()
		if _, ok := expected.GetOwned(name); ok {
			continue
		}
		p.error(arg, PropertyDoesNotExist, name)
	}
}
func reportMissingMembers(p *Parser, expected Object, received *BracedExpression) {
	membersSet := map[string]bool{}
	for _, member := range expected.Embedded {
		membersSet[member.Name] = true
	}
	for _, member := range expected.Members {
		membersSet[member.Name] = true
	}
	for _, member := range received.Expr.(*TupleExpression).Elements {
		if named, ok := member.(*Entry); ok && named.Key != nil {
			delete(membersSet, named.Key.(*Identifier).Text())
		}
	}

	if len(membersSet) == 0 {
		return
	}
	var msg string
	var i int
	for member := range membersSet {
		msg += fmt.Sprintf("'%v'", member)
		if i != len(membersSet)-1 {
			msg += ", "
		}
		i++
	}
	p.error(received, MissingKeys, msg)
}

func checkOptionInstanciation(p *Parser, i *InstanceExpression, t TypeAlias) {
	i.typing = t
	args := i.Args.Expr.(*TupleExpression)
	if len(args.Elements) == 0 {
		return
	}
	checkOptionArgs(p, args)
	expected := t.Params[0].Value
	received := args.Elements[0].Type()
	if !expected.Extends(received) {
		p.error(args.Elements[0], CannotAssignType, expected, received)
	}
}
func checkInferredOptionInstance(p *Parser, i *InstanceExpression) {
	args := i.Args.Expr.(*TupleExpression)
	if len(args.Elements) == 0 {
		i.typing = makeOptionType(Invalid{})
		p.error(i, MissingTypeArgs)
		return
	}
	checkOptionArgs(p, i.Args.Expr.(*TupleExpression))
	i.typing = makeOptionType(args.Elements[0].Type())
}
func checkOptionArgs(p *Parser, args *TupleExpression) {
	if len(args.Elements) > 1 {
		p.error(args, TooManyElements, 1, len(args.Elements))
	}
	for _, arg := range args.Elements {
		arg.typeCheck(p)
		switch arg.(type) {
		case *Param, *Entry:
			p.error(arg, InvalidPattern)
		}
	}
}

func typeCheckMapInstanciation(p *Parser, i *InstanceExpression, t TypeAlias) {
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	typeCheckTypeArgs(p, nil, t.Params)
	args := i.Args.Expr.(*TupleExpression).Elements
	formatMapEntries(p, args)
	i.typing = getMapType(p, i)
	typeCheckMapEntries(p, args, i.typing.(TypeAlias).Ref.(Map))
}
func checkInferredMapInstance(p *Parser, i *InstanceExpression) {
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	args := i.Args.Expr.(*TupleExpression).Elements
	formatMapEntries(p, args)
	if len(args) == 0 {
		p.error(i, MissingTypeArgs)
		i.typing = makeMapType(Invalid{}, Invalid{})
		return
	}
	i.typing = getInferredMapType(p, i)
	typeCheckMapEntries(p, args, i.typing.(TypeAlias).Ref.(Map))
}

func formatMapEntries(p *Parser, received []Expression) {
	for i := range received {
		received[i] = getFormattedMapEntry(p, received[i])
	}
}
func getFormattedMapEntry(p *Parser, received Expression) *Entry {
	entry, ok := received.(*Entry)
	if !ok {
		if param, ok := received.(*Param); ok {
			p.error(param, FieldExpected)
			return &Entry{Key: param.Identifier, Value: param.Complement}
		} else {
			p.error(received, FieldExpected)
			return &Entry{Value: received}
		}
	}
	if _, ok := entry.Key.(*Identifier); ok {
		p.error(entry.Key, FieldKeyExpected)
		entry.Key = nil
	}
	return entry
}
func getMapType(p *Parser, i *InstanceExpression) ExpressionType {
	t := i.Typing.Type().(Type).Value.(TypeAlias)
	args := i.Args.Expr.(*TupleExpression).Elements
	var key, value ExpressionType
	var kk, vk bool
	if len(args) > 0 {
		key, kk = t.Params[0].build(p.scope, args[0].(*Entry).Key.Type())
	} else {
		key, kk = t.Params[0].build(p.scope, nil)
	}
	if len(args) > 0 {
		value, vk = t.Params[1].build(p.scope, args[0].(*Entry).Value.Type())
	} else {
		value, vk = t.Params[1].build(p.scope, nil)
	}
	if !kk || !vk {
		p.error(i, MissingTypeArgs)
	}
	t.Params = append(t.Params[:0:0], t.Params...)
	t.Params[0].Value = key
	t.Params[1].Value = value
	m := t.Ref.(Map)
	m.Key = key
	m.Value = value
	t.Ref = m
	return t
}
func getInferredMapType(p *Parser, i *InstanceExpression) ExpressionType {
	args := i.Args.Expr.(*TupleExpression).Elements
	var key, value ExpressionType
	var kk, vk bool
	key, kk = args[0].(*Entry).Key.Type().build(p.scope, nil)
	value, vk = args[0].(*Entry).Value.Type().build(p.scope, nil)
	if !kk || !vk {
		p.error(i, MissingTypeArgs)
	}
	built, _ := makeMapType(key, value).build(p.scope, nil)
	return built
}
func typeCheckMapEntries(p *Parser, entries []Expression, t Map) {
	for i := range entries {
		entry := entries[i].(*Entry)
		if entry.Key != nil {
			entry.Key.typeCheck(p)
			var key ExpressionType
			if b, ok := entry.Key.(*BracketedExpression); ok {
				key = b.Expr.Type()
			} else {
				key = entry.Key.Type()
			}
			if !t.Key.Extends(key) {
				p.error(entry.Key, CannotAssignType, t.Key, key)
			}
		}
		if entry.Value != nil {
			entry.Value.typeCheck(p)
			if !t.Value.Extends(entry.Value.Type()) {
				p.error(entry.Key, CannotAssignType, t.Value, entry.Value)
			}
		}
	}
}

func typeCheckNamedListInstanciation(p *Parser, i *InstanceExpression) {
	// next line ensured by calling function
	typing := i.Typing.Type().(Type).Value.(List)
	i.typing = typing
	elements := i.Args.Expr.(*TupleExpression).Elements
	if len(elements) == 0 {
		return
	}
	first := elements[0]
	first.typeCheck(p)
	el := typing.Element
	if alias, ok := el.(TypeAlias); ok {
		p.pushScope(NewScope(ProgramScope))
		defer p.dropScope()
		typeCheckTypeArgs(p, nil, alias.Params)
		el, _ = el.build(p.scope, first.Type())
	}

	for i := range elements[1:] {
		elements[i+1].typeCheck(p)
		if !el.Extends(elements[i+1].Type()) {
			p.error(elements[i+1], CannotAssignType, el, elements[i+1].Type())
		}
	}
}

// Type-check nodes like []{a, b, c}
func checkInferredListInstance(p *Parser, i *InstanceExpression) {
	// next line ensured by calling function
	elements := i.Args.Expr.(*TupleExpression).Elements
	if len(elements) == 0 {
		p.error(i, MissingTypeArgs)
		i.typing = List{Invalid{}}
		return
	}
	first := elements[0]
	first.typeCheck(p)
	i.typing = List{first.Type()}

	for j := range elements[1:] {
		elements[j+1].typeCheck(p)
		if !i.typing.Extends(elements[j+1].Type()) {
			p.error(elements[j+1], CannotAssignType, i.typing, elements[j+1].Type())
		}
	}
}

func (i *InstanceExpression) Loc() Loc {
	return Loc{
		Start: i.Typing.Loc().Start,
		End:   i.Args.loc.End,
	}
}

func (p *Parser) parseInstanceExpression() Expression {
	expr := parseBinaryType(p)
	if p.Peek().Kind() != LeftBrace {
		return expr
	}
	args := p.parseBracedExpression()
	args.Expr = makeTuple(args.Expr)
	return &InstanceExpression{
		Typing: expr,
		Args:   args,
	}
}

func parseInferredInstance(p *Parser, constructor Expression) *InstanceExpression {
	args := p.parseBracedExpression()
	args.Expr = makeTuple(args.Expr)
	return &InstanceExpression{Typing: constructor, Args: args}
}
