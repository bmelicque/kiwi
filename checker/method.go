package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type Receiver struct {
	Name   Identifier
	Typing Identifier
}

type MethodDeclaration struct {
	Receiver    Receiver
	Name        Identifier
	Initializer Expression
	loc         tokenizer.Loc
}

func (m MethodDeclaration) Loc() tokenizer.Loc { return m.loc }

func (c *Checker) checkMethodDeclarationReceiver(expr parser.Node) (Receiver, bool) {
	tuple, ok := expr.(parser.TupleExpression)
	if !ok || len(tuple.Elements) != 1 {
		return Receiver{}, false
	}
	typed, ok := tuple.Elements[0].(parser.TypedExpression)
	if !ok {
		return Receiver{}, false
	}
	param := c.checkParam(typed)
	typing, ok := param.Typing.(*Identifier)
	if !ok {
		return Receiver{}, false
	}
	return Receiver{param.Identifier, *typing}, true
}
func (c *Checker) checkMethodDeclarationName(expr parser.Node) *Identifier {
	token, ok := expr.(parser.TokenExpression)
	if !ok {
		return nil
	}
	identifier, _ := c.checkToken(token, false).(*Identifier)
	return identifier
}
func (c *Checker) checkMethodDeclarationFunction(receiver Receiver, expr parser.Node) Expression {
	scope := NewShadowScope()
	scope.Add(receiver.Name.Token.Text(), receiver.Name.Loc(), receiver.Typing.Type().(Type).Value)
	c.pushScope(scope)
	defer c.dropScope()

	return c.CheckExpression(expr)
}

func (c *Checker) checkMethodDeclaration(a parser.Assignment) MethodDeclaration {
	left := a.Declared.(*parser.PropertyAccessExpression)

	start := left.Expr.Loc().Start
	receiver, ok := c.checkMethodDeclarationReceiver(left.Expr)
	if !ok {
		c.report("Expected receiver argument", left.Expr.Loc())
	}

	identifier := c.checkMethodDeclarationName(left.Property)
	if identifier == nil {
		c.report("Expected method name", left.Property.Loc())
		identifier = &Identifier{}
	}

	init := c.checkMethodDeclarationFunction(receiver, a.Initializer)
	if _, ok := init.Type().(Function); !ok {
		c.report("Expected function type", a.Initializer.Loc())
	}

	return MethodDeclaration{
		Receiver:    receiver,
		Name:        *identifier,
		Initializer: init,
		loc:         tokenizer.Loc{Start: start, End: init.Loc().End},
	}
}
