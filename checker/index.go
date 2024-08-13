package checker

import (
	"fmt"
	"reflect"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type CheckerError parser.ParserError

func (c CheckerError) Error() string { return c.Message }

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
	case parser.BinaryExpression:
		return c.checkBinaryExpression(node)
	case parser.CallExpression:
		return c.checkCallExpression(node)
	case parser.FunctionExpression:
		return c.checkFunctionExpression(node)
	case parser.ListExpression:
		return c.checkListExpression(node)
	case parser.ObjectDefinition:
		return c.checkObjectDefinition(node)
	case parser.ObjectExpression:
		return c.checkObjectExpression(node)
	case *parser.PropertyAccessExpression:
		return c.checkPropertyAccess(*node)
	case parser.RangeExpression:
		return c.checkRangeExpression(node)
	case *parser.TokenExpression:
		return c.checkToken(node, true)
	case parser.TupleExpression:
		return c.checkTuple(node)

	}
	panic(fmt.Sprintf("Cannot check type '%v' (not implemented yet)", reflect.TypeOf(node)))
}

func (c *Checker) Check(node parser.Node) Node {
	switch node := node.(type) {
	case parser.Assignment:
		operator := node.Operator.Kind()
		if operator != tokenizer.DECLARE && operator != tokenizer.DEFINE {
			return c.checkAssignment(node)
		}
		if _, ok := node.Declared.(*parser.PropertyAccessExpression); ok {
			return c.checkMethodDeclaration(node)
		}
		return c.checkVariableDeclaration(node)
	case parser.Body:
		return c.checkBody(node)
	case parser.ExpressionStatement:
		return c.checkExpressionStatement(node)
	case parser.For:
		return c.checkLoop(node)
	case parser.IfElse:
		return c.checkIf(node)
	case parser.Return:
		return c.checkReturnStatement(node)
	default:
		return c.CheckExpression(node)
	}
}
