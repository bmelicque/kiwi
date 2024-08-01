package parser

import (
	"fmt"

	"github.com/bmelicque/test-parser/tokenizer"
)

type ExpressionStatement struct {
	Expr Expression
}

func (s ExpressionStatement) Loc() tokenizer.Loc { return s.Expr.Loc() }
func (s ExpressionStatement) Check(c *Checker) {
	if s.Expr != nil {
		s.Expr.Check(c)
	}
}

type Assignment struct {
	Declared    Expression // "value", "Type", "(value Type).method", "(value Type).(..Trait)"
	Initializer Expression
	Typing      Expression
	Operator    tokenizer.Token
}

func (a Assignment) Loc() tokenizer.Loc {
	loc := a.Operator.Loc()
	if a.Declared != nil {
		loc.Start = a.Declared.Loc().Start
	} else if a.Typing != nil {
		loc.Start = a.Typing.Loc().Start
	}
	if a.Initializer != nil {
		loc.End = a.Initializer.Loc().End
	}
	return loc
}

func checkDeclarationTyping(c *Checker, a Assignment) {
	if a.Typing == nil {
		return
	}

	typing, ok := a.Typing.Type(c.scope).(Type)
	if !ok {
		c.report("Typing expected", a.Typing.Loc())
	} else if !typing.value.Extends(a.Initializer.Type(c.scope)) {
		c.report("Initializer type does not match declared type", a.Loc())
	}
}

func getDeclarationTyping(c *Checker, a Assignment) ExpressionType {
	if a.Typing == nil {
		return a.Initializer.Type(c.scope)
	}
	return a.Typing.Type(c.scope).(Type).value
}

func checkDeclaration(c *Checker, a Assignment) {
	checkDeclarationTyping(c, a)

	switch declared := a.Declared.(type) {
	case TokenExpression:
		if declared.Token.Kind() != tokenizer.IDENTIFIER {
			c.report("Identifier expected", declared.Loc())
			break
		}
		name := declared.Token.Text()
		c.scope.Add(name, declared.Loc(), getDeclarationTyping(c, a))
		a.Initializer.Check(c)
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
		a.Initializer.Check(c)
	case PropertyAccessExpression:
		tuple, ok := declared.Expr.(TupleExpression)
		if !ok || len(tuple.Elements) != 1 {
			c.report("Receiver argument expected", declared.Expr.Loc())
			break
		}
		expr, ok := tuple.Elements[0].(TypedExpression)
		if !ok {
			c.report("Receiver argument expected", declared.Expr.Loc())
			break
		}
		name, ok := CheckTypedIdentifier(c, expr)
		if !ok {
			c.report("Receiver argument expected", declared.Expr.Loc())
			break
		}

		// TODO: add to scope
		method, ok := declared.Property.(TokenExpression)
		if !ok || IsType(method) {
			c.report("Method name expected", declared.Property.Loc())
			break
		}

		scope := NewScope()
		scope.shadow = true
		scope.Add(name, expr.Loc(), expr.Typing.Type(c.scope))
		c.PushScope(scope)
		a.Initializer.Check(c)
		c.DropScope()

		function, ok := getDeclarationTyping(c, a).(Function)
		if !ok {
			c.report("Function type expected", a.Initializer.Loc())
			break
		}
		c.scope.AddMethod(method.Token.Text(), declared.Loc(), expr.Typing.Type(c.scope), function)
	default:
		c.report("Invalid pattern", declared.Loc())
		a.Initializer.Check(c)
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
		c.report("Types don't match", a.Loc())
	}
}

func checkAssignment(c *Checker, a Assignment) {
	a.Initializer.Check(c)

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
	// TODO: if type, only operator allowed is "::"
	if a.Initializer == nil {
		return
	}

	operator := a.Operator.Kind()
	if operator == tokenizer.DEFINE || operator == tokenizer.DECLARE || a.Typing != nil {
		checkDeclaration(c, a)
	} else if operator == tokenizer.ASSIGN {
		checkAssignment(c, a)
	}
}

func ParseAssignment(p *Parser) Statement {
	expr := ParseExpression(p)

	var typing Expression
	var operator tokenizer.Token
	var loc tokenizer.Loc
	next := p.tokenizer.Peek()
	switch next.Kind() {
	case tokenizer.COLON:
		p.tokenizer.Consume()
		typing = ParseExpression(p)
		operator = p.tokenizer.Consume()
		if operator.Kind() != tokenizer.ASSIGN {
			p.report("'=' expected", operator.Loc())
		}
	case tokenizer.DECLARE,
		tokenizer.DEFINE,
		tokenizer.ASSIGN:
		operator = p.tokenizer.Consume()
	default:
		return ExpressionStatement{expr}
	}
	init := ParseExpression(p)
	loc = operator.Loc()
	if expr != nil {
		loc.Start = expr.Loc().Start
	} else if typing != nil {
		loc.Start = typing.Loc().Start
	}
	if init != nil {
		loc.End = init.Loc().End
	}
	return Assignment{expr, init, typing, operator}
}
