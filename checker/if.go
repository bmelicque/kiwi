package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type If struct {
	Keyword   tokenizer.Token
	Condition Expression
	Body      Body
	Alternate Node // If | Body
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
	body := c.checkBody(*node.Body)
	c.dropScope()

	alternate := checkAlternate(c, node.Alternate)

	return If{
		Keyword:   node.Keyword,
		Condition: condition,
		Body:      body,
		Alternate: alternate,
	}
}

func checkAlternate(c *Checker, alternate parser.Node) Node {
	if alternate == nil {
		return nil
	}
	switch alternate := alternate.(type) {
	case parser.Body:
		return c.checkBody(alternate)
	case parser.IfElse:
		return c.checkIf(alternate)
	default:
		panic("Alternate should've been a block or an if-else!")
	}
}
