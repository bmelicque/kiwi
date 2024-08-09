package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type ListExpression struct {
	Elements []Expression
	loc      tokenizer.Loc
}

func (l ListExpression) Loc() tokenizer.Loc { return l.loc }

func (l ListExpression) Type() ExpressionType {
	if len(l.Elements) == 0 {
		return List{Primitive{UNKNOWN}}
	}
	t := l.Elements[0].Type()
	if t.Kind() == TYPE {
		return t
	}
	return List{t}
}

func (c *Checker) checkListExpression(node parser.ListExpression) ListExpression {
	var typing ExpressionType
	elements := make([]Expression, len(node.Elements))
	for i, element := range node.Elements {
		if element == nil {
			continue
		}
		elements[i] = c.CheckExpression(element)
		if typing == nil {
			typing = elements[i].Type()
		} else if !typing.Extends(elements[i].Type()) {
			c.report("Types don't match", element.Loc())
		}
	}
	return ListExpression{elements, node.Loc()}
}
