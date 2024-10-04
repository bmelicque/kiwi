package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type Exit struct {
	Operator tokenizer.Token
	Value    Expression // This may be nil
}

func (r Exit) Loc() tokenizer.Loc {
	loc := r.Operator.Loc()
	if r.Value != nil {
		loc.End = r.Value.Loc().End
	}
	return loc
}

func (c *Checker) checkExitStatement(statement parser.Exit) Exit {
	var value Expression
	if statement.Value != nil {
		value = c.checkExpression(statement.Value)
	}

	if statement.Operator.Kind() == tokenizer.RETURN_KW {
		checkReturnValue(c, statement, value)
	}

	return Exit{
		Operator: statement.Operator,
		Value:    value,
	}
}

func checkReturnValue(c *Checker, statement parser.Exit, value Expression) {
	expected := c.scope.returnType

	if expected == nil {
		if value != nil {
			c.report("No return value expected", value.Loc())
		}
		return
	}

	if value == nil {
		c.report("Return value expected", statement.Loc())
		return
	}

	if !expected.Extends(value.Type()) {
		c.report("Returned type doesn't match expected type", statement.Loc())
	}
}
