package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type ExpressionStatement struct {
	Expr Node
}

func (s ExpressionStatement) Loc() tokenizer.Loc { return s.Expr.Loc() }

type Assignment struct {
	Declared    Node // "value", "Type", "(value: Type).method"
	Initializer Node
	Typing      Node
	Operator    tokenizer.Token // '=', ':=', '::', '+='...
}

func (a Assignment) Loc() tokenizer.Loc {
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

func ParseAssignment(p *Parser) Node {
	expr := ParseExpression(p)

	var typing Node
	var operator tokenizer.Token
	var loc tokenizer.Loc
	next := p.tokenizer.Peek()
	switch next.Kind() {
	case tokenizer.COLON:
		p.tokenizer.Consume()
		typing = ParseExpression(p)
		operator = p.tokenizer.Consume()
		if operator.Kind() != tokenizer.ASSIGN {
			p.report("'=' expected", operator.Loc())
		}
	case tokenizer.DECLARE,
		tokenizer.DEFINE,
		tokenizer.ASSIGN:
		operator = p.tokenizer.Consume()
	default:
		return ExpressionStatement{expr}
	}
	init := ParseExpression(p)
	loc = operator.Loc()
	if expr != nil {
		loc.Start = expr.Loc().Start
	} else if typing != nil {
		loc.Start = typing.Loc().Start
	}
	if init != nil {
		loc.End = init.Loc().End
	}
	return Assignment{expr, init, typing, operator}
}
