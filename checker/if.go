package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type If struct {
	Keyword   tokenizer.Token
	Condition Expression
	Body      Body
}

func (c *Checker) checkIf(node parser.IfElse) If {
	condition := c.CheckExpression(node.Condition)
	if condition.Type().Kind() != BOOLEAN {
		c.report("Expected boolean condition", node.Condition.Loc())
	}

	scope := NewScope()
	scope.returnType = c.scope.returnType
	c.pushScope(scope)
	defer c.dropScope()
	body := c.checkBody(*node.Body)

	return If{
		Keyword:   node.Keyword,
		Condition: condition,
		Body:      body,
	}
}
