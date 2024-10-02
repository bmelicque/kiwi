package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func (c *Checker) checkDefinition(a parser.Assignment) Node {
	var pattern Expression
	constant := a.Operator.Kind() == tokenizer.DEFINE

	declared := a.Declared
	if d, ok := declared.(parser.ParenthesizedExpression); ok {
		declared = d.Expr
	}
	switch declared := declared.(type) {
	case parser.TokenExpression:
		return c.checkIdentifierDefinition(a)
	case parser.ComputedAccessExpression:
		return c.checkGenericTypeDefinition(a)
	case parser.PropertyAccessExpression:
		return c.checkMethodDeclaration(a)
	default:
		c.report("Invalid pattern", declared.Loc())
		init := c.checkExpression(a.Initializer)
		return VariableDeclaration{pattern, init, a.Loc(), constant}
	}
}

func (c *Checker) checkIdentifierDefinition(a parser.Assignment) VariableDeclaration {
	declared := a.Declared.(parser.TokenExpression)
	identifier, ok := c.checkToken(declared, false).(Identifier)
	if !ok {
		c.report("Identifier expected", declared.Loc())
		return VariableDeclaration{
			Pattern:     identifier,
			Initializer: c.checkExpression(a.Initializer),
			loc:         a.Loc(),
			Constant:    true,
		}
	}

	init := c.checkExpression(a.Initializer)
	if identifier.isType {
		c.declareType(identifier, init)
	} else {
		c.declareFunction(identifier, init)
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

// check a generic type definition, like:  Generic[TypeParam] :: { value TypeParam }
func (c *Checker) checkGenericTypeDefinition(a parser.Assignment) VariableDeclaration {
	declared := a.Declared.(parser.ComputedAccessExpression)
	identifier, ok := checkTypeIdentifier(c, declared.Expr)
	if !ok {
		c.report("Type identifier expected", declared.Expr.Loc())
	}

	var params Params
	if declared.Property.Expr != nil {
		params = c.checkTypeParams(declared.Property)
	}

	c.pushScope(NewScope())
	addTypeParamsToScope(c.scope, params)
	init := c.checkExpression(a.Initializer)
	c.dropScope()
	addGenericTypeToScope(c, identifier, params, init)

	return VariableDeclaration{
		Pattern:     ComputedAccessExpression{Expr: identifier, Property: params},
		Initializer: init,
		loc:         a.Loc(),
		Constant:    true,
	}
}
func addGenericTypeToScope(c *Checker, identifier Identifier, params Params, init Expression) {
	p := make([]Generic, len(params.Params))
	for i, param := range params.Params {
		p[i] = Generic{Name: param.Identifier.Text()}
	}

	t, ok := init.Type().(Type)
	if !ok {
		c.report("Type definition expected", init.Loc())
		return
	}

	if identifier.Text() != "" {
		c.scope.Add(identifier.Text(), identifier.Loc(), Type{TypeAlias{
			Name:   identifier.Text(),
			Params: p,
			Ref:    t.Value,
		}})
	}
}
