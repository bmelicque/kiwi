package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TokenExpression struct {
	Token tokenizer.Token
}

func (t TokenExpression) Type(ctx *Scope) ExpressionType {
	switch t.Token.Kind() {
	case tokenizer.NUMBER:
		return Primitive{NUMBER}
	case tokenizer.BOOLEAN:
		return Primitive{BOOLEAN}
	case tokenizer.STRING:
		return Primitive{STRING}
	case tokenizer.STR_KW:
		return Type{Primitive{STRING}}
	case tokenizer.NUM_KW:
		return Type{Primitive{NUMBER}}
	case tokenizer.BOOL_KW:
		return Type{Primitive{BOOLEAN}}
	case tokenizer.IDENTIFIER:
		variable, ok := ctx.Find(t.Token.Text())
		if !ok {
			break
		}
		if IsType(t) {
			typing := variable.typing.(Type)
			return TypeRef{
				name: t.Token.Text(),
				ref:  typing.value,
			}
		}
		return variable.typing
	}
	return Primitive{UNKNOWN}
}
func (t TokenExpression) Check(c *Checker) {
	if t.Token.Kind() == tokenizer.IDENTIFIER {
		c.scope.ReadAt(t.Token.Text(), t.Loc())
	}
}
func (t TokenExpression) Loc() tokenizer.Loc { return t.Token.Loc() }
func (TokenExpression) Parse(p *Parser) Expression {
	token := p.tokenizer.Consume()
	switch token.Kind() {
	case tokenizer.BOOLEAN, tokenizer.NUMBER, tokenizer.STRING, tokenizer.IDENTIFIER, tokenizer.BOOL_KW, tokenizer.NUM_KW, tokenizer.STR_KW:
		return TokenExpression{token}
	}
	p.report("Expression expected", token.Loc())
	return nil
}
