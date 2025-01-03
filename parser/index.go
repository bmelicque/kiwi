package parser

import (
	"io"
	"os"
	"path/filepath"
	"slices"
)

type Node interface {
	typeCheck(*Parser)
	Loc() Loc
	getChildren() []Node
}
type Expression interface {
	Node
	Type() ExpressionType
}

func fallback(p *Parser) Expression {
	switch p.Peek().Kind() {
	case Dot:
		p.Consume()
		return parseTraitExpression(p, nil)
	case LeftParenthesis:
		return p.parseFunctionExpression(nil)
	case LeftBrace:
		if p.allowBraceParsing {
			return p.parseBlock()
		}
	}
	return p.parseToken()
}

func Walk(node Node, predicate func(n Node, skip func())) {
	var s bool
	predicate(node, func() { s = true })
	if s {
		return
	}
	children := node.getChildren()
	for i := range children {
		Walk(children[i], predicate)
	}
}

func ParseProgram(reader io.Reader, path string) ([]Node, []ParserError, *Scope) {
	p := MakeParser(reader)
	p.filePath = path
	statements := []Node{}

	for p.Peek().Kind() != EOF {
		statements = append(statements, p.parseStatement())
		next := p.Peek().Kind()
		if next == EOL {
			p.DiscardLineBreaks()
		} else if next != EOF {
			p.error(&Literal{p.Peek()}, TokenExpected, token{kind: EOL})
		}
	}

	for i := range statements {
		statements[i].typeCheck(p)
	}

	if len(p.errors) > 0 {
		statements = []Node{}
	}
	return statements, p.errors, p.scope
}

var filesExports = map[string]Module{}

func ParseFile(path string) ([]Node, []ParserError) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	ast, errors, scope := ParseProgram(file, path)
	o := scope.toModule()
	filesExports[path] = o
	return ast, errors
}

type File struct {
	Path      string
	DependsOn []*File
	NeededFor []*File
}

func (f *File) hasAncestor(a *File) bool {
	for _, parent := range f.NeededFor {
		if parent == a || parent.hasAncestor(a) {
			return true
		}
	}
	return false
}

type DependencyBuilder struct {
	files   []*File
	inCycle []*File
}

func (d *DependencyBuilder) make() *DependencyBuilder {
	d.files = []*File{}
	d.inCycle = []*File{}
	return d
}

func (d *DependencyBuilder) makeFile(path string) *File {
	f := &File{
		Path:      path,
		DependsOn: []*File{},
		NeededFor: []*File{},
	}
	d.files = append(d.files, f)
	return f
}
func (d *DependencyBuilder) findFile(path string) *File {
	for _, file := range d.files {
		if file.Path == path {
			return file
		}
	}
	return nil
}
func (d *DependencyBuilder) buildDependencyTree(filePath string) *File {
	f := d.makeFile(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	dir := filepath.Dir(filePath)

	names := parseDependencies(file)
	files := make([]*File, len(names))
	i := 0
	for _, name := range names {
		name = filepath.Join(dir, name)
		found := d.findFile(name)
		if found != nil {
			files[i] = found
		} else {
			file := d.buildDependencyTree(name)
			if file == nil {
				continue
			}
			files[i] = file
		}
		files[i].NeededFor = append(files[i].NeededFor, f)
		i++
	}
	files = files[0:i]
	f.DependsOn = files
	return f
}

func (d *DependencyBuilder) validateDependencyTree() {
	for _, file := range d.files {
		if file.hasAncestor(file) {
			d.inCycle = append(d.inCycle, file)
		}
	}
}

func parseDependencies(reader io.Reader) []string {
	p := MakeParser(reader)
	files := []string{}

	for p.Peek().Kind() != EOF {
		statement := p.parseStatement()
		u, ok := statement.(*UseDirective)
		if !ok {
			break
		}
		if u.Source == nil {
			continue
		}
		source := u.Source.Text()
		files = append(files, source[1:len(source)-1]) // remove quotation marks
		next := p.Peek().Kind()
		if next == EOL {
			p.DiscardLineBreaks()
		} else if next != EOF {
			p.error(&Literal{p.Peek()}, TokenExpected, token{kind: EOL})
		}
	}

	return files
}

// returns pathsInOrder, pathsInCircularDependencies
func GetCompileOrder(rootPath string) ([]*File, []*File) {
	d := (&DependencyBuilder{}).make()
	d.buildDependencyTree(rootPath)
	d.validateDependencyTree()
	slices.Reverse(d.files)
	return d.files, d.inCycle
}

func Parse(rootPath string) ([][]Node, []ParserError) {
	files, _ := GetCompileOrder(rootPath)
	chunks := [][]Node{}
	errors := []ParserError{}
	for _, file := range files {
		ast, errs := ParseFile(file.Path)
		chunks = append(chunks, ast)
		errors = append(errors, errs...)
	}
	return chunks, errors
}
