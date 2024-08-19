package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type Return struct {
	Operator tokenizer.Token
	Value    Expression // This may be nil
}

func (r Return) Loc() tokenizer.Loc {
	loc := r.Operator.Loc()
	if r.Value != nil {
		loc.End = r.Value.Loc().End
	}
	return loc
}

func (c *Checker) checkReturnStatement(statement parser.Return) Return {
	expected := c.scope.returnType

	if expected == nil && statement.Value != nil {
		c.report("Expected no return value", statement.Loc())
	}
	if expected != nil && statement.Value == nil {
		c.report("Expected return value", statement.Loc())
	}

	var value Expression
	if statement.Value != nil {
		value = c.checkExpression(statement.Value)
	}
	return Return{
		Operator: statement.Operator,
		Value:    value,
	}
}
