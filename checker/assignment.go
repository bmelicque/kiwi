package checker

import "github.com/bmelicque/test-parser/parser"

type Assignment struct {
	Pattern  Expression
	Value    Expression
	Operator parser.Token
}

func (a Assignment) Loc() parser.Loc {
	loc := a.Operator.Loc()
	if a.Pattern != nil {
		loc.Start = a.Pattern.Loc().Start
	}
	if a.Value != nil {
		loc.End = a.Value.Loc().End
	}
	return loc
}

func (c *Checker) checkAssignment(assignment parser.Assignment) Assignment {
	pattern := c.checkExpression(assignment.Declared)
	value := c.checkExpression(assignment.Initializer)

	if !pattern.Type().Extends(value.Type()) {
		c.report("Types don't match", assignment.Loc())
		return Assignment{}
	}

	switch pattern := pattern.(type) {
	case Literal:
		c.report("Identifier expected", assignment.Declared.Loc())
	case Identifier:
		switch assignment.Operator.Kind() {
		case
			parser.AddAssign,
			parser.SubAssign,
			parser.MulAssign,
			parser.PowAssign,
			parser.DivAssign,
			parser.ModAssign:
			c.checkArithmetic(pattern, value)
		case parser.ConcatAssign:
			c.checkConcat(pattern, value)
		case
			parser.LogicalAndAssign,
			parser.LogicalOrAssign:
			c.checkLogical(pattern, value)
		}
	case TupleExpression:
		if assignment.Operator.Kind() != parser.Assign {
			c.report("Expected '='", assignment.Declared.Loc())
		}
		for _, element := range pattern.Elements {
			if _, ok := element.(Identifier); !ok {
				c.report("Expected identifier", element.Loc())
			}
		}
	}

	return Assignment{
		Pattern:  pattern,
		Value:    value,
		Operator: assignment.Operator,
	}
}
