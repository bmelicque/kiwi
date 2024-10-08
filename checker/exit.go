package checker

import "github.com/bmelicque/test-parser/parser"

type Exit struct {
	Operator parser.Token
	Value    Expression // This may be nil
}

func (r Exit) Loc() parser.Loc {
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
	if operator == parser.CONTINUE_KW && statement.Value != nil {
		c.report("No value expected after 'continue'", statement.Value.Loc())
	}
	if operator == parser.RETURN_KW && !c.scope.in(FunctionScope) {
		c.report("Cannot return outside of a function", statement.Loc())
	}
	if operator == parser.BREAK_KW && !c.scope.in(LoopScope) {
		c.report("Cannot break outside of a loop", statement.Loc())
	}
	if operator == parser.CONTINUE_KW && !c.scope.in(LoopScope) {
		c.report("Cannot continue outside of a loop", statement.Loc())
	}

	return Exit{
		Operator: statement.Operator,
		Value:    value,
	}
}
