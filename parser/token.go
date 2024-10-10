package parser

import (
	"fmt"
	"unicode"
)

type Literal struct {
	Token
}

func (l *Literal) typeCheck(_ *Parser) { return }

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
	isType bool
}

func (i *Identifier) typeCheck(p *Parser) {
	name := i.Text()
	if variable, ok := p.scope.Find(name); ok {
		p.scope.ReadAt(name, i.Loc())
		i.typing = variable.typing
	}
}

func (i *Identifier) Type() ExpressionType { return i.typing }

func (p *Parser) parseToken(expectNewName bool) Expression {
	token := p.Peek()
	switch token.Kind() {
	case BooleanLiteral, NumberLiteral, StringLiteral, BooleanKeyword, NumberKeyword, StringKeyword:
		p.Consume()
		return &Literal{token}
	case Name:
		p.Consume()
		isType := unicode.IsUpper(rune(token.Text()[0]))
		return &Identifier{Token: token, isType: isType}
	}
	if !p.allowEmptyExpr {
		p.Consume()
		p.report("Expression expected", token.Loc())
	}
	return nil
}
