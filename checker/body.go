package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type Body struct {
	Statements []Node
	loc        tokenizer.Loc
}

func (c *Checker) checkBody(body parser.Body) Body {
	var ended bool
	var unreachable tokenizer.Loc

	statements := make([]Node, len(body.Statements))
	for i, node := range body.Statements {
		statement := c.CheckExpression(node)
		statements[i] = statement

		if isEndStatement(statement) && !ended {
			ended = true
			unreachable.Start = node.Loc().Start
		}
		if ended {
			unreachable.End = node.Loc().End
		}
	}
	if ended {
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
