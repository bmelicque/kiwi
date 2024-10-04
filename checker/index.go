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
	scope := NewScope(ProgramScope)
	scope.outer = &std
	return &Checker{errors: []CheckerError{}, scope: scope}
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

func (c *Checker) checkExpression(node parser.Node) Expression {
	switch node := node.(type) {
	case parser.BinaryExpression:
		return c.checkBinaryExpression(node)
	case parser.Block:
		return c.checkBlock(node)
	case parser.CallExpression:
		return c.checkCallExpression(node)
	case parser.FunctionExpression:
		return c.checkFunctionExpression(node)
	case parser.ListTypeExpression:
		return c.checkListTypeExpression(node)
	case parser.ParenthesizedExpression:
		return c.checkParenthesizedExpression(node)
	case parser.PropertyAccessExpression:
		return c.checkPropertyAccess(node)
	case parser.RangeExpression:
		return c.checkRangeExpression(node)
	case parser.TokenExpression:
		return c.checkToken(node, true)
	case parser.TupleExpression:
		return c.checkTuple(node)
	case parser.TypedExpression:
		return c.checkParam(node)
	case parser.UnaryExpression:
		return c.checkUnaryExpression(node)
	}
	panic(fmt.Sprintf("Cannot check type '%v' (not implemented yet)", reflect.TypeOf(node)))
}

func (c *Checker) Check(node parser.Node) Node {
	switch node := node.(type) {
	case parser.Assignment:
		operator := node.Operator.Kind()
		if operator == tokenizer.DEFINE {
			return c.checkDefinition(node)
		}
		if operator != tokenizer.DECLARE {
			return c.checkAssignment(node)
		}
		return c.checkVariableDeclaration(node)
	case parser.ExpressionStatement:
		return c.checkExpressionStatement(node)
	case parser.ForExpression:
		return c.checkLoop(node)
	case parser.IfElse:
		return c.checkIf(node)
	case parser.MatchExpression:
		return c.checkMatchExpression(node)
	case parser.Exit:
		return c.checkExitStatement(node)
	default:
		return c.checkExpression(node)
	}
}
