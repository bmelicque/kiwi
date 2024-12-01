package parser

import (
	"fmt"
	"slices"
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
	i.Typing.typeCheck(p)
	t, ok := i.Typing.Type().(Type)
	if !ok {
		p.report("Type expected", i.Typing.Loc())
		i.Args.typeCheck(p)
		return
	}
	switch t.Value.(type) {
	case TypeAlias:
		typeCheckStructInstanciation(p, i)
	case List:
		typeCheckListInstanciation(p, i)
	default:
		p.report("This type cannot be instantiated", i.Typing.Loc())
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
	object, ok := alias.Ref.(Object)
	if !ok {
		p.report("Object type expected", i.Typing.Loc())
		i.typing = Unknown{}
		return
	}
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()
	typeCheckTypeArgs(p, nil, alias.Params)

	args := i.Args.Expr.(*TupleExpression).Elements
	formatStructEntries(p, args)
	for _, arg := range args {
		entry := arg.(*Entry)
		var name string
		if entry.Key != nil {
			name = entry.Key.(*Identifier).Text()
		}
		expected := object.Members[name]
		if !expected.Extends(entry.Value.Type()) {
			p.report("Type doesn't match the expected one", arg.Loc())
		}
	}

	reportExcessMembers(p, object.Members, args)
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
			p.report("':' expected between key and value", param.Loc())
			return &Entry{Key: param.Identifier, Value: param.Complement}
		} else {
			p.report("Entry expected", received.Loc())
			return &Entry{Value: received}
		}
	}
	if _, ok := entry.Key.(*Identifier); !ok {
		p.report("Identifier expected", entry.Key.Loc())
		entry.Key = &Identifier{}
	}
	return entry
}
func reportExcessMembers(p *Parser, expected map[string]ExpressionType, received []Expression) {
	for _, arg := range received {
		namedArg, ok := arg.(*Entry)
		if !ok {
			continue
		}
		name := namedArg.Key.(*Identifier).Text()
		if _, ok := expected[name]; ok {
			continue
		}
		p.report(
			fmt.Sprintf("Property '%v' doesn't exist on this type", name),
			arg.Loc(),
		)
	}
}
func reportMissingMembers(p *Parser, expected Object, received *BracedExpression) {
	membersSet := map[string]bool{}
	for name := range expected.Members {
		isOptional := slices.Contains(expected.Optionals, name)
		hasDefault := slices.Contains(expected.Defaults, name)
		if !isOptional && !hasDefault {
			membersSet[name] = true
		}
	}
	for _, member := range received.Expr.(*TupleExpression).Elements {
		if named, ok := member.(*Entry); ok {
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
	p.report(fmt.Sprintf("Missing key(s) %v", msg), received.loc)
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
			p.report("':' expected between key and value", param.Loc())
			return &Entry{Key: param.Identifier, Value: param.Complement}
		} else {
			p.report("Entry expected", received.Loc())
			return &Entry{Value: received}
		}
	}
	if _, ok := entry.Key.(*Identifier); ok {
		p.report("Literal or brackets expected", entry.Key.Loc())
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
		p.report(
			"Could not fully determine returned type, consider adding type arguments",
			i.Loc(),
		)
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
				p.report("Type doesn't match expected key type", entry.Key.Loc())
			}
		}
		if entry.Value != nil {
			entry.Value.typeCheck(p)
			if !t.Value.Extends(entry.Value.Type()) {
				p.report("Type doesn't match expected value type", entry.Value.Loc())
			}
		}
	}
}

func typeCheckListInstanciation(p *Parser, i *InstanceExpression) {
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
			p.report("Type doesn't match", elements[i+1].Loc())
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
