package parser

import (
	"fmt"
	"unicode"
)

type Literal struct {
	Token
}

func (l *Literal) typeCheck(_ *Parser) {}

func (l *Literal) Type() ExpressionType {
	switch l.Kind() {
	case NumberLiteral:
		return Primitive{NUMBER}
	case BooleanLiteral:
		return Primitive{BOOLEAN}
	case StringLiteral:
		return Primitive{STRING}
	case StringKeyword:
		return Type{Primitive{STRING}}
	case NumberKeyword:
		return Type{Primitive{NUMBER}}
	case BooleanKeyword:
		return Type{Primitive{BOOLEAN}}
	default:
		panic(fmt.Sprintf("Token kind '%v' not implemented yet", l.Kind()))
	}
}

type Identifier struct {
	Token
	typing ExpressionType
}

func (i *Identifier) IsType() bool {
	return unicode.IsUpper(rune(i.Token.Text()[0]))
}

func (i *Identifier) typeCheck(p *Parser) {
	name := i.Text()
	if variable, ok := p.scope.Find(name); ok {
		p.scope.ReadAt(name, i.Loc())
		i.typing = variable.typing
	} else {
		i.typing = Primitive{UNKNOWN}
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
		p.report("Expression expected", token.Loc())
	}
	return nil
}
