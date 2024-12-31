package main

import (
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
	source := os.Args[1]
	output := os.Args[2]

	file, err := os.Open(source)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// TODO:
	_, _ = parser.GetCompileOrder(source)

	ast, errors := parser.Parse(file)
	if len(errors) == 0 {
		f, err := os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		_, err = f.WriteString(emitter.EmitProgram(ast))
		if err != nil {
			log.Fatal(err)
		}
		f.Sync()
	} else {
		for _, err := range errors {
			line := err.Node.Loc().Start.Line
			col := err.Node.Loc().End.Col
			msg := err.Text()
			fmt.Printf("Error at line %v, col. %v: %v\n", line, col, msg)
		}
	}
}
