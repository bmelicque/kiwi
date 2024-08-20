package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type Assignment struct {
	Pattern  Expression
	Value    Expression
	Operator tokenizer.Token
}

func (a Assignment) Loc() tokenizer.Loc {
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
			tokenizer.ADD_ASSIGN,
			tokenizer.SUB_ASSIGN,
			tokenizer.MUL_ASSIGN,
			tokenizer.POW_ASSIGN,
			tokenizer.DIV_ASSIGN,
			tokenizer.MOD_ASSIGN:
			c.checkArithmetic(pattern, value)
		case tokenizer.CONCAT_ASSIGN:
			c.checkConcat(pattern, value)
		case
			tokenizer.LAND_ASSIGN,
			tokenizer.LOR_ASSIGN:
			c.checkLogical(pattern, value)
		}
	case TupleExpression:
		if assignment.Operator.Kind() != tokenizer.ASSIGN {
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
