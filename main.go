package main

import (
	"fmt"
	"log"

	"github.com/bmelicque/test-parser/checker"
	"github.com/bmelicque/test-parser/emitter"
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type TokenKind int

const (
	EOFToken TokenKind = iota
	BlankToken
	EOLToken
	DefinitionOperator
	DeclarationOperator
	AssignmentOperator
)

func main() {
	t, err := tokenizer.New("test.txt")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer t.Dispose()

	p := parser.MakeParser(t)
	c := checker.MakeChecker()
	program := p.ParseProgram()
	for _, statement := range program {
		c.Check(statement)
	}
	errors := append(p.GetReport(), c.GetReport()...)
	if len(errors) == 0 {
		e := emitter.MakeEmitter()
		for _, statement := range program {
			e.Emit(statement)
		}
		fmt.Printf("%+v\n", e.String())
	} else {
		for _, err := range errors {
			fmt.Printf("Error at line %v, col. %v: %v\n", err.Loc.Start.Line, err.Loc.Start.Col, err.Message)
		}
	}
}
