package parser

import (
	"fmt"

	"github.com/bmelicque/test-parser/tokenizer"
)

type ExpressionStatement struct {
	Expr Expression
}

func (s ExpressionStatement) Loc() tokenizer.Loc { return s.Expr.Loc() }
func (s ExpressionStatement) Check(c *Checker)   { s.Expr.Check(c) }

type Assignment struct {
	Declared    Expression // "value", "Type", "(value Type).method", "(value Type).(..Trait)"
	Initializer Expression
	typing      Expression
	Operator    tokenizer.Token
	loc         tokenizer.Loc
}

func (a Assignment) Loc() tokenizer.Loc { return a.loc }

func checkDeclaration(c *Checker, a Assignment) {
	switch declared := a.Declared.(type) {
	case TokenExpression:
		if declared.Token.Kind() != tokenizer.IDENTIFIER {
			c.report("Identifier expected", declared.Loc())
			break
		}
		name := declared.Token.Text()
		c.scope.Add(name, declared.Loc(), a.Initializer.Type(c.scope))
	case TupleExpression:
		typing, ok := a.Initializer.Type(c.scope).(Tuple)
		if !ok {
			c.report("Patterns don't match", a.Loc())
			break
		}
		if len(declared.Elements) > len(typing.elements) {
			c.report("Too many elements", declared.Loc())
			break
		}
		for i, expr := range declared.Elements {
			identifier, ok := expr.(TokenExpression)
			if !ok || identifier.Token.Kind() != tokenizer.IDENTIFIER {
				c.report("Identifier expected", identifier.Loc())
				continue
			}
			name := identifier.Token.Text()
			c.scope.Add(name, identifier.Loc(), typing.elements[i])
		}
	}
}

func handleIdentiferAssignment(c *Checker, expr Expression, typing ExpressionType, a Assignment) {
	token, ok := expr.(TokenExpression)
	if !ok || token.Token.Kind() != tokenizer.IDENTIFIER {
		c.report("Identifier expected", token.Loc())
		return
	}
	name := token.Token.Text()
	variable, ok := c.scope.Find(name)
	if !ok {
		c.report(fmt.Sprintf("'%v' not defined", name), token.Loc())
		return
	}
	c.scope.WriteAt(name, token.Loc())
	if !variable.typing.Extends(typing) {
		c.report("Types don't match", a.loc)
	}
}

func checkAssignment(c *Checker, a Assignment) {
	switch assignee := a.Declared.(type) {
	case TokenExpression:
		handleIdentiferAssignment(c, assignee, a.Initializer.Type(c.scope), a)
	case TupleExpression:
		typing, ok := a.Initializer.Type(c.scope).(Tuple)
		if !ok {
			c.report("Patterns don't match", a.Loc())
			break
		}
		if len(assignee.Elements) > len(typing.elements) {
			c.report("Too many elements", assignee.Loc())
			break
		}
		for i, expr := range assignee.Elements {
			handleIdentiferAssignment(c, expr, typing.elements[i], a)
		}
	}
}

func (a Assignment) Check(c *Checker) {
	a.Initializer.Check(c)
	switch a.Operator.Kind() {
	case tokenizer.DEFINE, tokenizer.DECLARE:
		checkDeclaration(c, a)
	case tokenizer.ASSIGN:
		checkAssignment(c, a)
	}
}

func ParseAssignment(p *Parser) Statement {
	expr := ParseExpression(p)

	var init Expression
	var operator tokenizer.Token
	var loc tokenizer.Loc
	next := p.tokenizer.Peek()
	switch next.Kind() {
	case tokenizer.COLON:
		// TODO: value: type = init
	case tokenizer.DECLARE,
		tokenizer.DEFINE,
		tokenizer.ASSIGN:
		operator = p.tokenizer.Consume()
		init = ParseExpression(p)
		// FIXME: handle expr == nil && init == nil
		loc = tokenizer.Loc{Start: expr.Loc().Start, End: init.Loc().End}
		return Assignment{expr, init, nil, operator, loc}
	}
	return ExpressionStatement{expr}
}
