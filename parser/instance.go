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
	if l, ok := i.Typing.(*ListTypeExpression); ok && l.Expr == nil {
		typeCheckAnonymousListInstanciation(p, i)
		return
	}
	i.Typing.typeCheck(p)
	t, ok := i.Typing.Type().(Type)
	if !ok {
		p.error(i.Typing, TypeExpected)
		i.Args.typeCheck(p)
		return
	}
	switch t.Value.(type) {
	case TypeAlias:
		typeCheckStructInstanciation(p, i)
	case List:
		typeCheckNamedListInstanciation(p, i)
	default:
		p.error(i.Typing, NotInstanceable)
	}
}

// Parse a struct instanciation, like 'Object{key: value}'
func typeCheckStructInstanciation(p *Parser, i *InstanceExpression) {
	// next line should be ensured by calling function
	alias := i.Typing.Type().(Type).Value.(TypeAlias)
	if alias.Name == "Map" {
		typeCheckMapInstanciation(p, i, alias)
		return
	}
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	typeCheckTypeArgs(p, nil, alias.Params)

	object, ok := alias.Ref.(Object)
	if !ok {
		p.error(i.Typing, ObjectTypeExpected)
		i.typing = Unknown{}
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

func typeCheckMapInstanciation(p *Parser, i *InstanceExpression, t TypeAlias) {
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	typeCheckTypeArgs(p, nil, t.Params)
	args := i.Args.Expr.(*TupleExpression).Elements
	formatMapEntries(p, args)
	i.typing = getMapType(p, i)
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
func typeCheckAnonymousListInstanciation(p *Parser, i *InstanceExpression) {
	// next line ensured by calling function
	elements := i.Args.Expr.(*TupleExpression).Elements
	if len(elements) == 0 {
		p.error(i, MissingTypeArgs)
		i.typing = List{Unknown{}}
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
	expr := p.parseUnaryExpression()
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
