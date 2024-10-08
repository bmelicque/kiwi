package checker

import "github.com/bmelicque/test-parser/parser"

type Block struct {
	Statements []Node
	loc        parser.Loc
}

func (b Block) Loc() parser.Loc { return b.loc }
func (b Block) reportLoc() parser.Loc {
	if len(b.Statements) > 0 {
		return b.Statements[len(b.Statements)-1].Loc()
	} else {
		return b.loc
	}
}
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
	var unreachable parser.Loc
	for _, statement := range statements {
		if foundExit {
			if unreachable.Start == (parser.Position{}) {
				unreachable.Start = statement.Loc().Start
			}
			unreachable.End = statement.Loc().End
		}
		if _, ok := statement.(Exit); ok {
			foundExit = true
		}
	}
	if unreachable != (parser.Loc{}) {
		c.report("Detected unreachable code", unreachable)
	}
}
