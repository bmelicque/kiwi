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

	operator := statement.Operator.Kind()
	if operator == tokenizer.RETURN_KW {
		checkReturnValue(c, statement, value)
	}
	if operator == tokenizer.CONTINUE_KW && statement.Value != nil {
		c.report("No value expected after 'continue'", statement.Value.Loc())
	}

	if operator == tokenizer.RETURN_KW && !c.scope.in(FunctionScope) {
		c.report("Cannot return outside of a function", statement.Loc())
	}
	if operator == tokenizer.BREAK_KW && !c.scope.in(LoopScope) {
		c.report("Cannot break outside of a loop", statement.Loc())
	}
	if operator == tokenizer.CONTINUE_KW && !c.scope.in(LoopScope) {
		c.report("Cannot continue outside of a loop", statement.Loc())
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
