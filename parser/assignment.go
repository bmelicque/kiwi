package parser

type ExpressionStatement struct {
	Expr Node
}

func (s ExpressionStatement) Loc() Loc { return s.Expr.Loc() }

type Assignment struct {
	Declared    Node // "value", "Type", "(value: Type).method"
	Initializer Node
	Typing      Node
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

func (p *Parser) parseAssignment() Node {
	expr := ParseExpression(p)

	var typing Node
	var operator Token
	next := p.Peek()
	switch next.Kind() {
	case COLON:
		p.Consume()
		typing = ParseExpression(p)
		operator = p.Consume()
		if operator.Kind() != ASSIGN {
			p.report("'=' expected", operator.Loc())
		}
	case DECLARE,
		DEFINE,
		ASSIGN:
		operator = p.Consume()
	default:
		return ExpressionStatement{expr}
	}
	init := ParseExpression(p)
	return Assignment{expr, init, typing, operator}
}
