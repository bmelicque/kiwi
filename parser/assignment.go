package parser

import "fmt"

type Assignment struct {
	Pattern  Expression // "value", "Type", "(value: Type).method"
	Value    Expression
	Operator Token // '=', ':=', '::', '+='...
}

func (a *Assignment) typeCheck(p *Parser) {
	switch a.Operator.Kind() {
	case Assign:
		typeCheckAssignment(p, a)
	case Declare:
		typeCheckDeclaration(p, a)
	case Define:
		typeCheckDefinition(p, a)
	default:
		panic("Assignment type check should've been exhaustive!")
	}
}

func (a *Assignment) Loc() Loc {
	loc := a.Operator.Loc()
	if a.Pattern != nil {
		loc.Start = a.Pattern.Loc().Start
	}
	if a.Value != nil {
		loc.End = a.Value.Loc().End
	}
	return loc
}

func (a *Assignment) getChildren() []Node {
	children := []Node{}
	if a.Pattern != nil {
		children = append(children, a.Pattern)
	}
	if a.Value != nil {
		children = append(children, a.Value)
	}
	return children
}

func (p *Parser) parseAssignment() Node {
	expr := p.parseExpression()

	var operator Token
	next := p.Peek()
	switch next.Kind() {
	case Declare,
		Define,
		Assign:
		operator = p.Consume()
	default:
		return expr
	}
	init := p.parseExpression()
	if c, ok := expr.(*ComputedAccessExpression); ok && operator.Kind() == Define {
		c.Property.Expr = makeTuple(c.Property.Expr)
		validateTypeParams(p, c.Property)
		expr = c
	}
	if b, ok := init.(*Block); ok && operator.Kind() == Define {
		validateObject(p, b)
		init = b
	}
	return &Assignment{expr, init, operator}
}

func validateObject(p *Parser, b *Block) {
	for i := range b.Statements {
		b.Statements[i] = getValidatedObjectField(p, b.Statements[i])
	}
}

func getValidatedObjectField(p *Parser, node Node) Node {
	switch node := node.(type) {
	case *Param:
		return node
	case *Entry:
		if _, ok := node.Key.(*Identifier); !ok {
			p.report("Identifier expected", node.Key.Loc())
			return nil
		}
		return node
	default:
		p.report("Field expected", node.Loc())
		return nil
	}
}

// type check assignment where operator is '='
func typeCheckAssignment(p *Parser, a *Assignment) {
	outer := p.writing
	p.writing = a
	a.Pattern.typeCheck(p)
	p.writing = outer
	a.Value.typeCheck(p)
	reportInvalidVariableType(p, a.Value)

	switch pattern := a.Pattern.(type) {
	case *ComputedAccessExpression:
		if isMap(pattern.Expr.Type()) {
			t := pattern.typing.(TypeAlias).Ref.(Sum).getMember("Some")
			if !t.Extends(a.Value.Type()) {
				p.report("Mismatched types", a.Loc())
			}
			return
		}
		if isSlice(pattern.Expr.Type()) {
			el := pattern.Expr.Type().(Ref).To.(List).Element
			if !el.Extends(a.Value.Type()) {
				p.report("Mismatched types", a.Loc())
			}
			return
		}
		p.report(
			"Invalid type, expected assignment to map element",
			pattern.Loc(),
		)
	case *Identifier:
		if pattern.typing.Extends(a.Value.Type()) {
			return
		}
		p.report(
			fmt.Sprintf(
				"Cannot assign value to '%v' (types don't match)",
				pattern.Text(),
			),
			pattern.Loc(),
		)
	case *TupleExpression:
		for _, element := range pattern.Elements {
			if _, ok := element.(*Identifier); !ok {
				p.report("Expected identifier", element.Loc())
			}
		}
		if !pattern.typing.Extends(a.Value.Type()) {
			p.report("Type doesn't match assignee's type", pattern.Loc())
		}
	default:
		p.report("Invalid pattern for assignment", a.Pattern.Loc())
	}
}

// type check assignment where operator is ':='
func typeCheckDeclaration(p *Parser, a *Assignment) {
	a.Value.typeCheck(p)
	reportInvalidVariableType(p, a.Value)
	switch pattern := a.Pattern.(type) {
	case *Identifier:
		declareIdentifier(p, pattern, a.Value.Type())
	case *TupleExpression:
		declareTuple(p, pattern, a.Value.Type())
	case *CallExpression:
		if !p.conditionalDeclaration {
			p.report("Invalid pattern", a.Pattern.Loc())
			return
		}
		p.typeCheckPattern(a.Pattern, a.Value.Type())
	default:
		p.report("Invalid pattern", a.Pattern.Loc())
	}
}

func declareIdentifier(p *Parser, identifier *Identifier, typing ExpressionType) {
	name := identifier.Text()
	if name == "" || name == "_" {
		return
	}
	if name == "Map" {
		msg := fmt.Sprintf("'%v' is a reserved name", name)
		p.report(msg, identifier.Loc())
		return
	}
	p.scope.Add(name, identifier.Loc(), typing)
}

func declareTuple(p *Parser, pattern *TupleExpression, typing ExpressionType) {
	tuple, ok := typing.(Tuple)
	if !ok {
		p.report(
			"Initializer type doesn't match pattern (expected tuple)",
			pattern.Loc(),
		)
		return
	}
	l := len(pattern.Elements)
	if l > len(tuple.Elements) {
		start := pattern.Elements[len(tuple.Elements)-1].Loc().Start
		end := pattern.Elements[l-1].Loc().End
		p.report("Too many elements", Loc{start, end})
		l = len(tuple.Elements)
	}
	for i := 0; i < l; i++ {
		identifier, ok := pattern.Elements[i].(*Identifier)
		if !ok {
			p.report("Identifier expected", pattern.Elements[i].Loc())
			continue
		}
		declareIdentifier(p, identifier, tuple.Elements[i])
	}
}

func typeCheckDefinition(p *Parser, a *Assignment) {
	switch pattern := a.Pattern.(type) {
	case *ComputedAccessExpression:
		typeCheckGenericTypeDefinition(p, a)
	case *Identifier:
		if a.Value == nil {
			return
		}
		a.Value.typeCheck(p)
		reportInvalidVariableType(p, a.Value)
		if pattern.IsType() {
			typeCheckTypeDefinition(p, a)
		} else {
			typeCheckFunctionDefinition(p, a)
		}
	case *PropertyAccessExpression:
		typeCheckMethod(p, pattern, a.Value)
	default:
		a.Value.typeCheck(p)
		reportInvalidVariableType(p, a.Value)
		p.report("Invalid pattern", pattern.Loc())
	}
}

func typeCheckGenericTypeDefinition(p *Parser, a *Assignment) {
	pattern := a.Pattern.(*ComputedAccessExpression)

	p.pushScope(NewScope(ProgramScope))
	typeCheckTypeParams(p, pattern.Property)
	a.Value.typeCheck(p)
	p.dropScope()

	identifier, ok := pattern.Expr.(*Identifier)
	if !ok || !identifier.IsType() {
		p.report("Type identifier expected", pattern.Expr.Loc())
		return
	}
	t := Type{TypeAlias{
		Name:   identifier.Text(),
		Params: pattern.Property.getGenerics(),
		Ref:    getInitType(p, a.Value),
	}}
	p.scope.Add(identifier.Text(), pattern.Loc(), t)
}

func typeCheckTypeDefinition(p *Parser, a *Assignment) {
	identifier := a.Pattern.(*Identifier)
	t := Type{TypeAlias{
		Name: identifier.Text(),
		Ref:  getInitType(p, a.Value),
	}}
	declareIdentifier(p, identifier, t)
}

func getInitType(p *Parser, expr Expression) ExpressionType {
	if expr == nil {
		return Unknown{}
	}
	if b, ok := expr.(*Block); ok {
		return getObjectDefinedType(p, b)
	}
	t, ok := expr.Type().(Type)
	if !ok {
		p.report("Type expected", expr.Loc())
		return Unknown{}
	}
	return t.Value
}

func getObjectDefinedType(p *Parser, b *Block) ExpressionType {
	o := newObject()
	for _, s := range b.Statements {
		switch s := s.(type) {
		case *Param:
			key := s.Identifier.Text()
			t, ok := s.Complement.Type().(Type)
			if !ok {
				p.report("Type expected, got value", s.Complement.Loc())
				o.Members[key] = Unknown{}
				continue
			}
			if isOptionType(t) {
				o.Optionals = append(o.Optionals, key)
			}
			o.Members[key] = t.Value
		case *Entry:
			key := s.Key.(*Identifier).Text()
			if _, ok := s.Value.Type().(Type); ok {
				p.report("Value expected, got type", s.Value.Loc())
			}
			o.Defaults = append(o.Defaults, key)
			o.Members[key] = s.Value.Type()
		}
	}
	return o
}

func typeCheckFunctionDefinition(p *Parser, a *Assignment) {
	identifier := a.Pattern.(*Identifier)
	t, ok := a.Value.Type().(Function)
	if !ok {
		p.report("Function expected", a.Value.Loc())
		return
	}
	declareIdentifier(p, identifier, t)
}

func typeCheckMethod(p *Parser, expr *PropertyAccessExpression, init Expression) {
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()

	typeIdentifier := declareMethodReceiver(p, expr.Expr)

	method, ok := expr.Property.(*Identifier)
	if !ok {
		p.report("Identifier expected", expr.Property.Loc())
	}

	init.typeCheck(p)

	if _, ok := init.Type().(Function); !ok {
		p.report("Function expected", init.Loc())
		return
	}
	if !ok || typeIdentifier == nil {
		return
	}

	t, ok := typeIdentifier.Type().(Type)
	if !ok {
		return
	}
	alias, ok := t.Value.(TypeAlias)
	if !ok {
		return
	}

	p.scope.AddMethod(method.Text(), alias, init.Type().(Function))
}

func declareMethodReceiver(p *Parser, receiver Expression) *Identifier {
	paren, ok := receiver.(*ParenthesizedExpression)
	if !ok || paren.Expr == nil {
		p.report("Receiver argument expected", receiver.Loc())
		return nil
	}
	param, ok := paren.Expr.(*Param)
	if !ok {
		p.report("Receiver argument expected", paren.Expr.Loc())
		return nil
	}

	typeIdentifier, ok := param.Complement.(*Identifier)
	if !ok || !typeIdentifier.IsType() {
		p.report("Type identifier expected", param.Complement.Loc())
		return nil
	}
	typeIdentifier.typeCheck(p)

	if t, ok := typeIdentifier.Type().(Type); ok {
		p.scope.Add(
			param.Identifier.Text(),
			param.Identifier.Loc(),
			t.Value,
		)
	}
	return typeIdentifier
}

func reportInvalidVariableType(p *Parser, value Expression) {
	switch t := value.Type().(type) {
	case TypeAlias:
		if t.Name != "!" {
			return
		}
		p.report(
			"Cannot declare variable as Result, consider using a 'try' or 'catch' expression",
			value.Loc(),
		)
	case Nil:
		p.report(
			"Cannot declare variables as nil, consider using option type",
			value.Loc(),
		)
	}
}
