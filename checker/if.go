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

func (i If) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: i.Keyword.Loc().Start,
		End:   i.Body.Loc().End,
	}
}

func (c *Checker) checkIf(node parser.IfElse) If {
	condition := c.checkExpression(node.Condition)
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
