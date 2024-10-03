package checker

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type TupleExpression struct {
	Elements []Expression
	typing   ExpressionType
	loc      tokenizer.Loc
}

func (t TupleExpression) Loc() tokenizer.Loc   { return t.loc }
func (t TupleExpression) Type() ExpressionType { return t.typing }

func (c *Checker) checkTuple(tuple parser.TupleExpression) TupleExpression {
	elements := checkTupleElements(c, tuple)
	checkTupleElementsConsistency(c, elements)
	checkTupleElementDuplicates(c, elements)
	return TupleExpression{elements, getTupleType(elements), tuple.Loc()}
}

func checkTupleElements(c *Checker, tuple parser.TupleExpression) []Expression {
	elements := make([]Expression, len(tuple.Elements))
	for i, element := range tuple.Elements {
		if node, ok := element.(parser.TypedExpression); ok {
			elements[i] = c.checkParam(node)
		} else {
			elements[i] = c.checkExpression(element)
		}
	}
	return elements
}

func checkTupleElementsConsistency(c *Checker, elements []Expression) {
	if len(elements) == 0 {
		return
	}
	_, expectParams := elements[0].(Param)
	for _, el := range elements[1:] {
		_, isParam := el.(Param)
		if expectParams && !isParam {
			c.report("Typed param expected", el.Loc())
		}
		if !expectParams && isParam {
			c.report("No type expected", el.Loc())
		}
	}
}

func checkTupleElementDuplicates(c *Checker, elements []Expression) {
	if len(elements) == 0 {
		return
	}
	params := map[string][]tokenizer.Loc{}
	for _, el := range elements {
		param, ok := el.(Param)
		if !ok {
			continue
		}
		name := param.Identifier.Text()
		params[name] = append(params[name], param.Identifier.Loc())
	}
	for name, locs := range params {
		if len(locs) == 1 {
			continue
		}
		for _, loc := range locs {
			c.report(fmt.Sprintf("Duplicate element '%v'", name), loc)
		}
	}
}

func getTupleType(elements []Expression) ExpressionType {
	if len(elements) == 0 {
		return Primitive{UNKNOWN}
	}
	if _, ok := elements[0].(Param); ok {
		return getTupleObjectType(elements)
	}
	if len(elements) == 1 {
		return elements[0].Type()
	}
	types := make([]ExpressionType, len(elements))
	for i, element := range elements {
		types[i] = element.Type()
	}
	return Tuple{types}
}

func getTupleObjectType(elements []Expression) ExpressionType {
	value := Object{map[string]ExpressionType{}}
	for _, element := range elements {
		param, ok := element.(Param)
		if !ok {
			continue
		}
		t, ok := param.Complement.Type().(Type)
		if !ok {
			continue
		}
		value.Members[param.Identifier.Text()] = t.Value
	}
	return Type{value}
}
