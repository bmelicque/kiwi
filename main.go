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
	parsed := p.ParseProgram()
	program := make([]checker.Node, len(parsed))
	for i, statement := range parsed {
		program[i] = c.Check(statement)
	}
	parserErrors := p.GetReport()
	checkerErrors := c.GetReport()
	if len(parserErrors)+len(checkerErrors) == 0 {
		e := emitter.MakeEmitter()
		for _, statement := range program {
			e.Emit(statement)
		}
		fmt.Printf("%+v\n", e.String())
	} else {
		for _, err := range parserErrors {
			fmt.Printf("Error at line %v, col. %v: %v\n", err.Loc.Start.Line, err.Loc.Start.Col, err.Message)
		}
		for _, err := range checkerErrors {
			fmt.Printf("Error at line %v, col. %v: %v\n", err.Loc.Start.Line, err.Loc.Start.Col, err.Message)
		}
	}
}
