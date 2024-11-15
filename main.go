package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

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
	file, err := os.Open("test.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	p, err := parser.MakeParser(bufio.NewReader(file))
	if err != nil {
		log.Fatal(err)
		return
	}

	parsed := p.ParseProgram()
	program := make([]parser.Node, len(parsed))
	parserErrors := p.GetReport()
	if len(parserErrors) == 0 {
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
	}
}
