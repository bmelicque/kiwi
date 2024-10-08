package checker

import "github.com/bmelicque/test-parser/parser"

type ListTypeExpression struct {
	Expr  Expression
	start parser.Position
}

func (a ListTypeExpression) Loc() parser.Loc {
	return parser.Loc{Start: a.start, End: a.Expr.Loc().End}
}

func (a ListTypeExpression) Type() ExpressionType {
	t, ok := a.Expr.Type().(Type)
	if !ok {
		return Type{List{Primitive{UNKNOWN}}}
	}
	return Type{List{t.Value}}
}

func (c *Checker) checkListTypeExpression(list parser.ListTypeExpression) ListTypeExpression {
	if list.Bracketed.Expr != nil {
		c.report("No expression expected", list.Bracketed.Loc())
	}

	expr := c.checkExpression(list.Type)
	if expr != nil && expr.Type().Kind() != TYPE {
		c.report("Type expected", expr.Loc())
	}

	return ListTypeExpression{expr, list.Bracketed.Loc().Start}
}
