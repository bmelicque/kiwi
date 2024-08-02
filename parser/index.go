package parser

import "github.com/bmelicque/test-parser/tokenizer"

type Node interface {
	Loc() tokenizer.Loc
}

type Typing interface {
	Node
	Type(ctx *Scope) ExpressionType
}

type Expression interface {
	Node
	Type() ExpressionType
	Check(c *Checker)
}

type Statement interface {
	Node
	Check(p *Checker)
}
