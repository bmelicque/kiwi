package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bmelicque/test-parser/emitter"
	"github.com/bmelicque/test-parser/parser"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
	parser.Program
	path string
}

func transpile(rootPath string, outDir string) {
	outDir = filepath.Dir(outDir)
	emptyOutDir(outDir)
	files, _ := parser.GetCompileOrder(rootPath)
	htmlPath := filepath.Join(filepath.Dir(rootPath), "index.html")
	h := parseHtml(htmlPath)
	outPath := rootPath[:len(rootPath)-len(filepath.Ext(rootPath))] + ".js"
	appendScript(h, filepath.Base(outPath))
	emitHtml(outDir, h)

	chunks := []chunk{}
	errors := []parser.ParserError{}
	for _, f := range files {
		program, errs := parser.ParseFile(f.Path)
		chunks = append(chunks, chunk{
			path:    getOutPath(rootPath, f.Path, outDir),
			Program: program,
		})
		errors = append(errors, errs...)
	}

	if len(errors) > 0 {
		logErrors(errors)
		return
	}

	std := emitter.EmitStd(filepath.Dir(rootPath), outDir)
	std = filepath.Join(outDir, std)
	for _, chunk := range chunks {
		writeChunk(chunk, std)
	}
}

func parseHtml(path string) *html.Node {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	doc, err := html.Parse(f)
	if err != nil {
		fmt.Println("Cannot parse index.html: ", err)
		return nil
	}
	return doc
}

func appendScript(n *html.Node, outName string) {
	h := findHtmlNode(n, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "head"
	})
	if h == nil {
		fmt.Println("Could not find head")
		return
	}
	scriptNode := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Lookup([]byte("script")),
		Data:     "script",
		Attr: []html.Attribute{
			{Key: "type", Val: "module"},
			{Key: "crossorigin"},
			{Key: "src", Val: outName},
		},
	}
	h.AppendChild(scriptNode)
}

// TODO: remove recursivity (use ParentNode if no NextSibling, beware no to go to parent of starting node)
func findHtmlNode(n *html.Node, predicate func(*html.Node) bool) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if predicate(c) {
			return c
		}
		if n := findHtmlNode(c, predicate); n != nil {
			return n
		}
	}
	return nil
}

func emitHtml(outDir string, n *html.Node) {
	f, err := os.Create(filepath.Join(outDir, "index.html"))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// TODO: handle error?
	html.Render(f, n)
	f.Sync()
}

func getOutPath(rootPath, filePath, outDir string) string {
	relative, _ := filepath.Rel(filepath.Dir(rootPath), filePath)
	outFile := filepath.Join(outDir, relative)
	ext := len(filepath.Ext(outFile))
	return outFile[:len(outFile)-ext] + ".js"
}

func writeChunk(chunk chunk, stdPath string) {
	f, err := os.Create(chunk.path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	stdPath, _ = filepath.Rel(filepath.Dir(chunk.path), stdPath)
	stdPath = stdPath[:len(stdPath)-3]
	stdPath = filepath.ToSlash(stdPath)
	if stdPath[0] != '.' {
		stdPath = "./" + stdPath
	}
	f.WriteString("import * as __ from \"" + stdPath + ".js\";\n")
	_, err = f.WriteString(emitter.EmitProgram(chunk.Program))
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

func emptyOutDir(dir string) {
	d, err := os.Open(dir)
	if err != nil {
		panic(err)
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		panic(err)
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			panic(err)
		}
	}
}
