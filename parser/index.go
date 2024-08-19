package parser

import "github.com/bmelicque/test-parser/tokenizer"

type Node interface {
	Loc() tokenizer.Loc
}
