package parser

import "github.com/bmelicque/test-parser/tokenizer"

type TypedExpression struct {
	expr   Expression
	typing Expression
}

func (t TypedExpression) Type(ctx *Scope) ExpressionType {
	// FIXME:
	return Primitive{UNKNOWN}
}
func (t TypedExpression) Emit(e *Emitter) { t.expr.Emit(e) }
func (t TypedExpression) Check(c *Checker) {
	switch expr := t.expr.(type) {
	case TokenExpression:
		if expr.Token.Kind() != tokenizer.IDENTIFIER {
			c.report("Identifer expected", expr.Loc())
		}
	default:
		c.report("Identifer expected", expr.Loc())
	}

	if t.typing.Type(c.scope).Kind() != TYPE {
		c.report("Type expected", t.typing.Loc())
	}
}
func (t TypedExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{Start: t.expr.Loc().Start, End: t.typing.Loc().End}
}

func ParseTypedExpression(p *Parser) Expression {
	expr := ParseExpression(p)

	if p.tokenizer.Peek().Kind() != tokenizer.COLON {
		return expr
	}

	p.tokenizer.Consume()
	typing := ParseExpression(p)
	return TypedExpression{expr, typing}
}
