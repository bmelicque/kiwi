package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type ComputedAccessExpression struct {
	Expr     Expression
	Property Expression
	typing   ExpressionType
}

func (c ComputedAccessExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: c.Expr.Loc().Start,
		End:   c.Property.Loc().End,
	}
}
func (c ComputedAccessExpression) Type() ExpressionType {
	return c.typing
}

func (c *Checker) checkComputedAccessExpression(node parser.ComputedAccessExpression) ComputedAccessExpression {
	expr := c.checkExpression(node.Expr)
	prop := checkBracketed(c, &node.Property)

	var typing ExpressionType
	switch t := expr.Type().(type) {
	case Type:
		// Generics
		alias, ok := t.Value.(TypeAlias)
		if !ok {
			c.report("No type arguments expected", node.Property.Loc())
			typing = Primitive{UNKNOWN}
			break
		}
		c.pushScope(NewScope(ProgramScope))
		defer c.dropScope()
		params := append(alias.Params[:0:0], alias.Params...)
		c.addTypeArgsToScope(prop, params)
		ref, _ := alias.Ref.build(c.scope, nil)
		typing = Type{TypeAlias{
			Name:   alias.Name,
			Params: params,
			Ref:    ref,
		}}
	case Function:
		// Generic function
		c.pushScope(NewScope(ProgramScope))
		defer c.dropScope()
		typeParams := append(t.TypeParams[:0:0], t.TypeParams...)
		c.addTypeArgsToScope(prop, typeParams)
		params := make([]ExpressionType, len(typeParams))
		for i, param := range typeParams {
			params[i], _ = param.build(c.scope, nil)
		}
		returned, _ := t.Returned.build(c.scope, nil)
		typing = Function{
			TypeParams: typeParams,
			Params:     Tuple{params},
			Returned:   returned,
		}
	case List:
		if prop.Type().Kind() != NUMBER {
			c.report("Number expected", prop.loc)
		}
		typing = t.Element
	}

	return ComputedAccessExpression{
		Expr:     expr,
		Property: prop,
		typing:   typing,
	}
}
