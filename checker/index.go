package checker

import (
	"fmt"
	"reflect"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type CheckerError parser.ParserError
type Node parser.Node
type Expression interface {
	Node
	Type() ExpressionType
}

type Checker struct {
	errors []CheckerError
	scope  *Scope
}

func (c *Checker) report(message string, loc tokenizer.Loc) {
	c.errors = append(c.errors, CheckerError{message, loc})
}
func (c Checker) GetReport() []CheckerError {
	return c.errors
}

func MakeChecker() *Checker {
	return &Checker{errors: []CheckerError{}, scope: NewScope()}
}

func (c *Checker) pushScope(scope *Scope) {
	scope.outer = c.scope
	c.scope = scope
}

func (c *Checker) dropScope() {
	for _, info := range c.scope.variables {
		if len(info.reads) == 0 {
			c.report("Unused variable", info.declaredAt)
		}
	}
	c.scope = c.scope.outer
}

func (c *Checker) CheckExpression(node parser.Node) Expression {
	switch node := node.(type) {
	case *parser.TokenExpression:
		return c.checkToken(node, true)
	}
	panic(fmt.Sprintf("Cannot check type '%v' (not implemented yet)", reflect.TypeOf(node)))
}

func (c *Checker) Check(node parser.Node) Node {
	switch node := node.(type) {
	case *parser.TokenExpression:
		return c.checkToken(node, true)
	}
	panic(fmt.Sprintf("Cannot check type '%v' (not implemented yet)", reflect.TypeOf(node)))
}
