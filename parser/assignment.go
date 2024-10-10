package parser

import "fmt"

type ExpressionStatement struct {
	Expr Expression
}

func (s ExpressionStatement) Loc() Loc { return s.Expr.Loc() }

type Assignment struct {
	Declared    Expression // "value", "Type", "(value: Type).method"
	Initializer Expression
	Typing      Expression
	Operator    Token // '=', ':=', '::', '+='...
}

func (a Assignment) Loc() Loc {
	loc := a.Operator.Loc()
	if a.Declared != nil {
		loc.Start = a.Declared.Loc().Start
	} else if a.Typing != nil {
		loc.Start = a.Typing.Loc().Start
	}
	if a.Initializer != nil {
		loc.End = a.Initializer.Loc().End
	}
	return loc
}

type VariableDeclaration struct {
	Pattern     Expression
	Initializer Expression
	loc         Loc
	Constant    bool
}

func (vd VariableDeclaration) Loc() Loc { return vd.loc }

func (p *Parser) parseAssignment() Node {
	expr := ParseExpression(p)

	var typing Expression
	var operator Token
	next := p.Peek()
	fmt.Printf("%#v\n", next)
	switch next.Kind() {
	case Colon:
		p.Consume()
		typing = ParseExpression(p)
		operator = p.Consume()
		if operator.Kind() != Assign {
			p.report("'=' expected", operator.Loc())
		}
	case Declare,
		Define,
		Assign:
		operator = p.Consume()
	default:
		return ExpressionStatement{expr}
	}
	init := ParseExpression(p)
	return Assignment{expr, init, typing, operator}
}
