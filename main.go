package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bmelicque/test-parser/checker"
	"github.com/bmelicque/test-parser/emitter"
	"github.com/bmelicque/test-parser/parser"
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
	t, err := parser.NewTokenizer("test.txt")
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
		f, err := os.Create("out.js")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		_, err = f.WriteString(emitter.EmitProgram(program))
		if err != nil {
			log.Fatal(err)
		}
		f.Sync()
	} else {
		for _, err := range parserErrors {
			fmt.Printf("Error at line %v, col. %v: %v\n", err.Loc.Start.Line, err.Loc.Start.Col, err.Message)
		}
		for _, err := range checkerErrors {
			fmt.Printf("Error at line %v, col. %v: %v\n", err.Loc.Start.Line, err.Loc.Start.Col, err.Message)
		}
	}
}
