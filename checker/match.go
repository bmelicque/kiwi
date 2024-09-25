package checker

import (
	"fmt"
	"slices"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type MatchCase struct {
	Pattern    Expression
	Typing     Identifier
	Statements []Node
}

type MatchStatement struct {
	Value Expression
	Cases []MatchCase
	loc   tokenizer.Loc
}

func (m MatchStatement) Loc() tokenizer.Loc { return m.loc }

var matchableType = []ExpressionTypeKind{SUM, TRAIT}

func (c *Checker) checkMatchStatement(node parser.MatchStatement) Node {
	value := checkMatchValue(c, node.Value)
	typing := checkMatchValueType(c, value)

	nodes := checkFirstCaseHasKeyword(c, node.Cases)
	if len(nodes) < 2 {
		c.report("At least 2 cases expected", node.Loc())
	}

	cases := []MatchCase{}
	for _, node := range nodes {
		cases = append(cases, checkMatchCase(c, node, typing))
	}

	return MatchStatement{value, cases, node.Loc()}
}

func checkMatchValue(c *Checker, value parser.Node) Expression {
	if value == nil {
		return nil
	}
	return c.checkExpression(value)
}
func checkMatchValueType(c *Checker, value Expression) ExpressionType {
	if value == nil {
		return nil
	}
	typing := value.Type()
	if alias, ok := typing.(TypeAlias); ok {
		typing = alias.Ref
	}
	if !slices.Contains(matchableType, typing.Kind()) {
		c.report("Invalid type (expected trait or sum type)", value.Loc())
	}
	return typing
}
func checkFirstCaseHasKeyword(c *Checker, cases []parser.MatchCase) []parser.MatchCase {
	if len(cases) == 0 {
		return nil
	}
	first := cases[0]
	if first.Pattern == nil {
		c.report("'case' expected", first.Statements[0].Loc())
		return cases[1:]
	}
	return cases
}
func checkMatchCase(c *Checker, node parser.MatchCase, matchedType ExpressionType) MatchCase {
	c.pushScope(NewScope())
	defer c.dropScope()

	pattern, identifier := checkPattern(c, node.Pattern, matchedType)
	statements := []Node{}
	for _, statement := range node.Statements {
		statements = append(statements, c.Check(statement))
	}
	return MatchCase{pattern, identifier, statements}
}
func checkPattern(c *Checker, pattern parser.Node, matchedType ExpressionType) (Expression, Identifier) {
	pattern = parser.Unwrap(pattern)

	var expr, typing parser.Node
	if t, ok := pattern.(parser.TypedExpression); ok {
		typing = t.Typing
		expr = t.Expr
	} else {
		typing = pattern
	}

	id, t := checkCaseTyping(c, typing, matchedType)
	var p Expression
	if expr != nil {
		if t == nil {
			c.report("Nothing expected (no constructor for this variant)", expr.Loc())
			t = Primitive{UNKNOWN}
		}
		p = declareCasePattern(c, expr, t)
	}

	return p, id
}
func checkCaseTyping(c *Checker, subtyping parser.Node, typing ExpressionType) (Identifier, ExpressionType) {
	token, ok := subtyping.(parser.TokenExpression)
	if !ok {
		c.report("Type identifier expected", subtyping.Loc())
		return Identifier{}, Primitive{UNKNOWN}
	}
	identifier, ok := c.checkToken(token, false).(Identifier)
	if !ok {
		c.report("Type identifier expected", subtyping.Loc())
		return Identifier{}, Primitive{UNKNOWN}
	}
	if !identifier.isType {
		c.report("Type identifier expected", subtyping.Loc())
	}

	name := identifier.Text()

	switch typing := typing.(type) {
	case Sum:
		member, ok := typing.Members[name]
		if !ok {
			c.report(fmt.Sprintf("Constructor '%v' doesn't exist on this type", name), identifier.Loc())
			return identifier, Primitive{UNKNOWN}
		}
		return identifier, member
	case Trait:
		v, _ := c.scope.Find(name)
		if v == nil {
			c.report(fmt.Sprintf("Cannot find type '%v'", name), identifier.Loc())
			return identifier, Primitive{UNKNOWN}
		}
		t := v.typing.(Type).Value
		alias, ok := t.(TypeAlias)
		if !ok || !alias.implements(typing) {
			c.report(fmt.Sprintf("Type '%v' doesn't implement this trait", name), identifier.Loc())
			return identifier, Primitive{UNKNOWN}
		}
		return identifier, alias
	default:
		panic("Case match not implemented yet!")
	}
}
func declareCasePattern(c *Checker, node parser.Node, t ExpressionType) Expression {
	node = parser.Unwrap(node)

	switch node := node.(type) {
	case parser.TokenExpression:
		identifier, ok := c.checkToken(node, false).(Identifier)
		if !ok {
			c.report("Identifier expected", node.Loc())
			return identifier
		}
		c.scope.Add(identifier.Text(), node.Loc(), t)
		return identifier
	case parser.TupleExpression, ListTypeExpression:
		panic("Case pattern not implemented yet!")
	default:
		c.report("Invalid pattern", node.Loc())
		return nil
	}
}
