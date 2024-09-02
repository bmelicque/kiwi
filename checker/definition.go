package checker

import (
	"unicode"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type TypeDefinition struct {
	Identifier  Identifier
	Initializer Expression
	loc         tokenizer.Loc
}

func (td TypeDefinition) Loc() tokenizer.Loc { return td.loc }

func (c *Checker) checkDefinition(a parser.Assignment) Node {
	var pattern Expression
	init := c.checkExpression(a.Initializer)
	constant := a.Operator.Kind() == tokenizer.DEFINE

	declared := a.Declared
	if d, ok := declared.(parser.ParenthesizedExpression); ok {
		declared = d.Expr
	}
	switch declared := declared.(type) {
	case parser.TokenExpression:
		return c.checkIdentifierDefinition(a)
	case parser.PropertyAccessExpression:
		return c.checkMethodDeclaration(a)
	default:
		c.report("Invalid pattern", declared.Loc())
	}
	return VariableDeclaration{pattern, init, a.Loc(), constant}
}

func (c *Checker) checkIdentifierDefinition(a parser.Assignment) VariableDeclaration {
	declared := a.Declared.(parser.TokenExpression)
	init := c.checkExpression(a.Initializer)

	identifier, ok := c.checkToken(declared, false).(Identifier)
	if !ok {
		c.report("Identifier expected", declared.Loc())
		return VariableDeclaration{
			Pattern:     identifier,
			Initializer: init,
			loc:         a.Loc(),
			Constant:    true,
		}
	}
	name := identifier.Token.Text()
	isTypeIdentifier := unicode.IsUpper(rune(name[0]))

	if !isTypeIdentifier {
		c.declareFunction(identifier, init)
	} else if t, ok := init.(GenericTypeDef); ok {
		c.declareGenericType(identifier, t)
	} else {
		c.declareType(identifier, init)
	}

	return VariableDeclaration{
		Pattern:     identifier,
		Initializer: init,
		loc:         a.Loc(),
		Constant:    true,
	}
}

func (c *Checker) declareType(identifier Identifier, init Expression) {
	t := init.Type()
	if tok, ok := t.(Type); ok {
		t = Type{TypeAlias{Name: identifier.Text(), Ref: tok.Value}}
	} else {
		c.report("Type expected", init.Loc())
	}
	c.scope.Add(identifier.Text(), identifier.Loc(), t)
}
func (c *Checker) declareFunction(identifier Identifier, init Expression) {
	t := init.Type()
	if t.Kind() != FUNCTION {
		c.report("Function type expected", init.Loc())
	}

	c.scope.Add(identifier.Text(), identifier.Loc(), t)
}
func (c *Checker) declareGenericType(identifier Identifier, init GenericTypeDef) {
	params := make([]ExpressionType, len(init.TypeParams.Params))
	for i, param := range init.TypeParams.Params {
		params[i] = param.Typing.Type()
	}

	t := TypeAlias{
		Name:   identifier.Text(),
		Params: params,
		Ref:    init.Expr.Type(),
	}
	c.scope.Add(identifier.Text(), identifier.Loc(), t)
}
