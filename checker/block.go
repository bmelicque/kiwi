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
	statements := make([]Node, len(block.Statements))
	for i, node := range block.Statements {
		statements[i] = c.Check(node)
	}
	reportUnreachableCode(c, statements)

	return Block{
		Statements: statements,
		loc:        block.Loc(),
	}
}

func reportUnreachableCode(c *Checker, statements []Node) {
	var foundExit bool
	var unreachable tokenizer.Loc
	for _, statement := range statements {
		if foundExit {
			if unreachable.Start == (tokenizer.Position{}) {
				unreachable.Start = statement.Loc().Start
			}
			unreachable.End = statement.Loc().End
		}
		if _, ok := statement.(Exit); ok {
			foundExit = true
		}
	}
	if unreachable != (tokenizer.Loc{}) {
		c.report("Detected unreachable code", unreachable)
	}
}
