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

func parseFile(reader io.Reader) ([]Node, []ParserError) {
	p := MakeParser(reader)
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
	return statements, p.errors
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
		panic(err)
	}
	defer file.Close()

	dir := filepath.Dir(filePath)

	names := parseDependencies(file)
	files := make([]*File, len(names))
	for i, name := range names {
		name = filepath.Join(dir, name)
		found := d.findFile(name)
		if found != nil {
			files[i] = found
		} else {
			files[i] = d.buildDependencyTree(name)
		}
		files[i].NeededFor = append(files[i].NeededFor, f)
	}
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

func Parse(reader io.Reader) ([]Node, []ParserError) {
	return parseFile(reader)
}
