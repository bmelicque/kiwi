package parser

import "fmt"

type Assignment struct {
	Pattern  Expression // "value", "Type", "(value: Type).method"
	Value    Expression
	Operator Token // '=', ':=', '::', '+='...
}

func (a *Assignment) typeCheck(p *Parser) {
	a.Value.typeCheck(p)
	switch a.Operator.Kind() {
	case Assign:
		typeCheckAssignment(p, a)
	case Declare:
		typeCheckDeclaration(p, a)
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

	switch pattern := a.Pattern.(type) {
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
	switch pattern := a.Pattern.(type) {
	case *Identifier:
		name := pattern.Text()
		if name == "" || name == "_" {
			return
		}
		p.scope.Add(name, pattern.Loc(), a.Value.Type())
	case *TupleExpression:
		// TODO: validate pattern declaration
	case *CallExpression:
		if !p.conditionalDeclaration {
			p.report("Invalid pattern", a.Pattern.Loc())
			return
		}
	default:
		p.report("Invalid pattern", a.Pattern.Loc())
	}
}
