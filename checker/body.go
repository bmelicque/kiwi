package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type Body struct {
	Statements []Node
	loc        tokenizer.Loc
}

func (b Body) Loc() tokenizer.Loc { return b.loc }

func (c *Checker) checkBody(body parser.Block) Body {
	var ended bool
	var unreachable tokenizer.Loc

	statements := make([]Node, len(body.Statements))
	for i, node := range body.Statements {
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

	return Body{
		Statements: statements,
		loc:        body.Loc(),
	}
}

func isEndStatement(statement Node) bool {
	_, ok := statement.(Return)
	return ok
}
