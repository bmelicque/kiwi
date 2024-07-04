package parser

import (
	"fmt"

	"github.com/bmelicque/test-parser/tokenizer"
)

type ExpressionStatement struct {
	expr Expression
}

func (s ExpressionStatement) Emit(e *Emitter) {
	s.expr.Emit(e)
	e.Write(";\n")
}
func (s ExpressionStatement) Loc() tokenizer.Loc { return s.expr.Loc() }
func (s ExpressionStatement) Check(c *Checker)   { s.expr.Check(c) }

type Assignment struct {
	declared    Expression // "value", "Type", "(value Type).method", "(value Type).(..Trait)"
	initializer Expression
	typing      Expression
	operator    tokenizer.Token
	loc         tokenizer.Loc
}

func (a Assignment) Emit(e *Emitter) {
	kind := a.operator.Kind()
	if kind == tokenizer.DEFINE {
		e.Write("const ")
	} else if kind == tokenizer.DECLARE {
		e.Write("let ")
	}
	a.declared.Emit(e)
	e.Write(" = ")
	a.initializer.Emit(e)
	if _, ok := a.initializer.(FunctionExpression); !ok {
		e.Write(";\n")
	}
}
func (a Assignment) Loc() tokenizer.Loc { return a.loc }

func checkDeclaration(c *Checker, a Assignment) {
	switch declared := a.declared.(type) {
	case TokenExpression:
		if declared.Token.Kind() != tokenizer.IDENTIFIER {
			c.report("Identifier expected", declared.Loc())
			break
		}
		name := declared.Token.Text()
		c.scope.Add(name, declared.Loc(), a.initializer.Type(c.scope))
	case TupleExpression:
		typing, ok := a.initializer.Type(c.scope).(Tuple)
		if !ok {
			c.report("Patterns don't match", a.Loc())
			break
		}
		if len(declared.elements) > len(typing.elements) {
			c.report("Too many elements", declared.Loc())
			break
		}
		for i, expr := range declared.elements {
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
	switch assignee := a.declared.(type) {
	case TokenExpression:
		handleIdentiferAssignment(c, assignee, a.initializer.Type(c.scope), a)
	case TupleExpression:
		typing, ok := a.initializer.Type(c.scope).(Tuple)
		if !ok {
			c.report("Patterns don't match", a.Loc())
			break
		}
		if len(assignee.elements) > len(typing.elements) {
			c.report("Too many elements", assignee.Loc())
			break
		}
		for i, expr := range assignee.elements {
			handleIdentiferAssignment(c, expr, typing.elements[i], a)
		}
	}
}

func (a Assignment) Check(c *Checker) {
	a.initializer.Check(c)
	switch a.operator.Kind() {
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
