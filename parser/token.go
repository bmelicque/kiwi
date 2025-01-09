package parser

import (
	"fmt"
	"unicode"
)

type Literal struct {
	Token
}

func (l *Literal) getChildren() []Node {
	return []Node{}
}

func (l *Literal) typeCheck(_ *Parser) {}

func (l *Literal) Type() ExpressionType {
	switch l.Kind() {
	case NumberLiteral:
		return Number{}
	case BooleanLiteral:
		return Boolean{}
	case StringLiteral:
		return String{}
	case StringKeyword:
		return Type{String{}}
	case NumberKeyword:
		return Type{Number{}}
	case BooleanKeyword:
		return Type{Boolean{}}
	default:
		panic(fmt.Sprintf("Token kind '%v' not implemented yet", l.Kind()))
	}
}

type Identifier struct {
	Token
	typing ExpressionType
	scope  *Scope
}

func (i *Identifier) getChildren() []Node {
	return []Node{}
}

func (i *Identifier) IsPrivate() bool {
	return i.Text()[0] == '_'
}
func (i *Identifier) IsType() bool {
	text := i.Text()
	var firstLetter rune
	if text[0] == '_' {
		if len(text) == 1 {
			return false
		}
		firstLetter = rune(text[1])
	} else {
		firstLetter = rune(text[0])
	}

	return unicode.IsUpper(firstLetter)
}
func (i *Identifier) GetScope() *Scope { return i.scope }

func (i *Identifier) typeCheck(p *Parser) {
	name := i.Text()
	if variable, ok := p.scope.Find(name); ok {
		if p.writing != nil {
			variable.writeAt(p.writing)
		} else {
			variable.readAt(i.Loc())
		}
		i.typing = variable.Typing
	} else {
		i.typing = Unknown{}
	}
}

func (i *Identifier) Type() ExpressionType { return i.typing }

func (p *Parser) parseToken() Expression {
	token := p.Peek()
	switch token.Kind() {
	case BooleanLiteral, NumberLiteral, StringLiteral, BooleanKeyword, NumberKeyword, StringKeyword:
		p.Consume()
		return &Literal{token}
	case Name:
		p.Consume()
		return &Identifier{Token: token}
	}
	if !p.allowEmptyExpr {
		p.error(&Literal{token}, ExpressionExpected)
	}
	return nil
}
