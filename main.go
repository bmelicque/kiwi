package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	outDir := os.Args[2]

	transpile(source, outDir)
}

type chunk struct {
	path  string
	nodes []parser.Node
}

func transpile(rootPath string, outDir string) {
	outDir = filepath.Dir(outDir)
	files, _ := parser.GetCompileOrder(rootPath)

	chunks := []chunk{}
	errors := []parser.ParserError{}
	for _, f := range files {
		nodes, errs := parser.ParseFile(f.Path)
		chunks = append(chunks, chunk{
			path:  getOutPath(rootPath, f.Path, outDir),
			nodes: nodes,
		})
		errors = append(errors, errs...)
	}

	if len(errors) > 0 {
		logErrors(errors)
		return
	}

	for _, chunk := range chunks {
		writeChunk(chunk)
	}
}

func getOutPath(rootPath, filePath, outDir string) string {
	relative, _ := filepath.Rel(filepath.Dir(rootPath), filePath)
	outFile := filepath.Join(outDir, relative)
	ext := len(filepath.Ext(outFile))
	return outFile[:len(outFile)-ext] + ".js"
}

func writeChunk(chunk chunk) {
	f, err := os.Create(chunk.path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.WriteString(emitter.EmitProgram(chunk.nodes))
	if err != nil {
		log.Fatal(err)
	}
	f.Sync()
}

func logErrors(errors []parser.ParserError) {
	for _, err := range errors {
		line := err.Node.Loc().Start.Line
		col := err.Node.Loc().Start.Col
		msg := err.Text()
		fmt.Printf("Error at line %v, col. %v: %v\n", line, col, msg)
	}
}
