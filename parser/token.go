package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TokenExpression struct {
	Token  tokenizer.Token
	typing ExpressionType
}

func (t *TokenExpression) setType(ctx *Scope) {
	switch t.Token.Kind() {
	case tokenizer.NUMBER:
		t.typing = Primitive{NUMBER}
	case tokenizer.BOOLEAN:
		t.typing = Primitive{BOOLEAN}
	case tokenizer.STRING:
		t.typing = Primitive{STRING}
	case tokenizer.STR_KW:
		t.typing = Type{Primitive{STRING}}
	case tokenizer.NUM_KW:
		t.typing = Type{Primitive{NUMBER}}
	case tokenizer.BOOL_KW:
		t.typing = Type{Primitive{BOOLEAN}}
	case tokenizer.IDENTIFIER:
		variable, ok := ctx.Find(t.Token.Text())
		if !ok {
			break
		}
		if IsType(t) {
			typing := variable.typing.(Type)
			t.typing = TypeRef{
				Name: t.Token.Text(),
				ref:  typing.value,
			}
		}
		t.typing = variable.typing
	default:
		t.typing = Primitive{UNKNOWN}
	}
}
func (t *TokenExpression) Type() ExpressionType { return t.typing }
func (t *TokenExpression) Check(c *Checker) {
	if t.Token.Kind() == tokenizer.IDENTIFIER {
		c.scope.ReadAt(t.Token.Text(), t.Loc())
	}
	if t.Type() == nil {
		t.setType(c.scope)
	}
}
func (t *TokenExpression) Loc() tokenizer.Loc { return t.Token.Loc() }
func (TokenExpression) Parse(p *Parser) Expression {
	token := p.tokenizer.Consume()
	switch token.Kind() {
	case tokenizer.BOOLEAN, tokenizer.NUMBER, tokenizer.STRING, tokenizer.IDENTIFIER, tokenizer.BOOL_KW, tokenizer.NUM_KW, tokenizer.STR_KW:
		return &TokenExpression{token, nil}
	}
	p.report("Expression expected", token.Loc())
	return nil
}
