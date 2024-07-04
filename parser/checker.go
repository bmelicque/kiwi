package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type Variable struct {
	declaredAt tokenizer.Loc
	typing     ExpressionType
	writes     []tokenizer.Loc
	reads      []tokenizer.Loc
}

type Scope struct {
	inner      map[string]*Variable
	returnType ExpressionType // The expected type for a return statement (if any) // TODO: handle no return vs optional return
	outer      *Scope
}

func (s Scope) Find(name string) (*Variable, bool) {
	variable, ok := s.inner[name]
	if ok {
		return variable, true
	}
	if s.outer != nil {
		return s.outer.Find(name)
	}
	return nil, false
}

func (s Scope) Has(name string) bool {
	_, ok := s.inner[name]
	return ok
}

func (s *Scope) Add(name string, declaredAt tokenizer.Loc, typing ExpressionType) {
	s.inner[name] = &Variable{declaredAt, typing, []tokenizer.Loc{}, []tokenizer.Loc{}}
}

func (s *Scope) WriteAt(name string, loc tokenizer.Loc) {
	variable, ok := s.Find(name)
	// TODO: panic on error
	if ok {
		variable.writes = append(variable.writes, loc)
	}
}

func (s *Scope) ReadAt(name string, loc tokenizer.Loc) {
	variable, ok := s.Find(name)
	// TODO: panic on error
	if ok {
		variable.reads = append(variable.reads, loc)
	}
}

func (s Scope) GetReturnType() ExpressionType {
	return s.returnType
}

type Checker struct {
	errors []ParserError
	scope  *Scope
}

func (c *Checker) report(message string, loc tokenizer.Loc) {
	c.errors = append(c.errors, ParserError{message, loc})
}
func (c Checker) GetReport() []ParserError {
	return c.errors
}

func MakeChecker() *Checker {
	return &Checker{errors: []ParserError{}, scope: &Scope{map[string]*Variable{}, nil, nil}}
}

func (c *Checker) PushScope(scope *Scope) {
	scope.outer = c.scope
	c.scope = scope
}

func (c *Checker) DropScope() {
	for _, info := range c.scope.inner {
		if len(info.reads) == 0 {
			c.report("Unused variable", info.declaredAt)
		}
	}
	c.scope = c.scope.outer
}
