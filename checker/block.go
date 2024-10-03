package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type Block struct {
	Statements []Node
	loc        tokenizer.Loc
}

func (b Block) Loc() tokenizer.Loc { return b.loc }
func (b Block) Type() ExpressionType {
	if len(b.Statements) == 0 {
		return Primitive{NIL}
	}
	last := b.Statements[len(b.Statements)-1]
	expr, ok := last.(Expression)
	if !ok {
		return Primitive{NIL}
	}
	return expr.Type()
}

func (c *Checker) checkBlock(block parser.Block) Block {
	var ended bool
	var unreachable tokenizer.Loc

	statements := make([]Node, len(block.Statements))
	for i, node := range block.Statements {
		statement := c.Check(node)
		statements[i] = statement

		if ended {
			if unreachable.Start == (tokenizer.Position{}) {
				unreachable.Start = node.Loc().Start
			}
			unreachable.End = node.Loc().End
		} else if isEndStatement(statement) {
			ended = true
		}
	}
	if unreachable != (tokenizer.Loc{}) {
		c.report("Detected unreachable code", unreachable)
	}

	return Block{
		Statements: statements,
		loc:        block.Loc(),
	}
}

func isEndStatement(statement Node) bool {
	_, ok := statement.(Return)
	return ok
}
