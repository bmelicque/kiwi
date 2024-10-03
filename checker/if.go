package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type If struct {
	Keyword   tokenizer.Token
	Condition Expression
	Block     Block
	Alternate Expression // If | Body
}

func (i If) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: i.Keyword.Loc().Start,
		End:   i.Block.Loc().End,
	}
}
func (i If) Type() ExpressionType { return i.Block.Type() }

func (c *Checker) checkIf(node parser.IfElse) If {
	condition := c.checkExpression(node.Condition)
	if condition.Type().Kind() != BOOLEAN {
		c.report("Expected boolean condition", node.Condition.Loc())
	}

	scope := NewScope()
	scope.returnType = c.scope.returnType
	c.pushScope(scope)
	block := c.checkBlock(*node.Body)
	c.dropScope()

	alternate := checkAlternate(c, node.Alternate)
	if alternate != nil && !block.Type().Extends(alternate.Type()) {
		c.report("Types of truthy and alternate blocks don't match", node.Loc())
	}

	return If{
		Keyword:   node.Keyword,
		Condition: condition,
		Block:     block,
		Alternate: alternate,
	}
}

func checkAlternate(c *Checker, alternate parser.Node) Expression {
	if alternate == nil {
		return nil
	}
	switch alternate := alternate.(type) {
	case parser.Block:
		return c.checkBlock(alternate)
	case parser.IfElse:
		return c.checkIf(alternate)
	default:
		panic("Alternate should've been a block or an if-else!")
	}
}
