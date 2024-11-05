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
	return &Assignment{expr, init, operator}
}

// type check assignment where operator is '='
func typeCheckAssignment(p *Parser, a *Assignment) {
	a.Pattern.typeCheck(p)
	a.Value.typeCheck(p)
	reportInvalidVariableType(p, a.Value)

	switch pattern := a.Pattern.(type) {
	case *ComputedAccessExpression:
		alias, ok := pattern.Expr.Type().(TypeAlias)
		if !ok || alias.Name != "Map" {
			p.report(
				"Invalid type, expected assignment to map element",
				pattern.Loc(),
			)
			return
		}
		t := pattern.typing.(TypeAlias).Ref.(Sum).getMember("Some")
		if !t.Extends(a.Value.Type()) {
			p.report("Mismatched types", a.Loc())
		}
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
	if name == "Map" || name == "Promise" {
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
	if l > len(tuple.elements) {
		start := pattern.Elements[len(tuple.elements)-1].Loc().Start
		end := pattern.Elements[l-1].Loc().End
		p.report("Too many elements", Loc{start, end})
		l = len(tuple.elements)
	}
	for i := 0; i < l; i++ {
		identifier, ok := pattern.Elements[i].(*Identifier)
		if !ok {
			p.report("Identifier expected", pattern.Elements[i].Loc())
			continue
		}
		declareIdentifier(p, identifier, tuple.elements[i])
	}
}

func typeCheckDefinition(p *Parser, a *Assignment) {
	switch pattern := a.Pattern.(type) {
	case *ComputedAccessExpression:
		p.pushScope(NewScope(ProgramScope))
		pattern.Property.Expr = makeTuple(pattern.Property.Expr)
		validateTypeParams(p, pattern.Property)
		typeCheckTypeParams(p, pattern.Property)
		a.Value.typeCheck(p)
		p.dropScope()

		identifier, ok := pattern.Expr.(*Identifier)
		if !ok || !identifier.IsType() {
			p.report("Type identifier expected", pattern.Expr.Loc())
			return
		}
		var ref ExpressionType
		if init, ok := a.Value.Type().(Type); ok {
			ref = init.Value
		} else {
			ref = Unknown{}
		}
		t := Type{TypeAlias{
			Name:   identifier.Text(),
			Params: pattern.Property.getGenerics(),
			Ref:    ref,
		}}
		p.scope.Add(identifier.Text(), pattern.Loc(), t)
	case *Identifier:
		a.Value.typeCheck(p)
		reportInvalidVariableType(p, a.Value)
		ok := true
		if !pattern.IsType() {
			p.report("Type identifier expected", pattern.Loc())
			ok = false
		}
		if a.Value != nil {
			if _, k := a.Value.Type().(Type); !k {
				p.report("Type expected", a.Value.Loc())
				ok = false
			}
		}
		if !ok {
			return
		}
		t := Type{TypeAlias{
			Name: pattern.Text(),
			Ref:  a.Value.Type().(Type).Value,
		}}
		declareIdentifier(p, pattern, t)
	case *PropertyAccessExpression:
		typeCheckMethod(p, pattern, a.Value)
	default:
		a.Value.typeCheck(p)
		reportInvalidVariableType(p, a.Value)
		p.report("Invalid pattern", pattern.Loc())
	}
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
