package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type FunctionExpression struct {
	Params   TupleExpression
	Operator tokenizer.Token // -> or =>
	Expr     Expression      // return value for '->', return type for '=>'
	Body     *Body
}

func (f FunctionExpression) Check(c *Checker) {
	functionScope := NewScope()

	for _, param := range f.Params.Elements {
		name, ok := CheckTypedIdentifier(c, param)
		if ok {
			typing := ReadTypeExpression(param.(TypedExpression).Typing)
			identifier := param.(TypedExpression).Expr.(*TokenExpression)
			functionScope.Add(name, identifier.Loc(), typing)
		}
	}

	if f.Operator.Kind() == tokenizer.SLIM_ARR {
		if f.Expr != nil {
			f.Expr.Check(c)
			if f.Expr.Type().Kind() == TYPE {
				c.report("Expression expected", f.Expr.Loc())
			}
		}

		if f.Body != nil {
			c.report("Function body should be a single statement", f.Body.Loc())
		}
		return
	}

	if f.Expr != nil {
		f.Expr.Check(c)
		if f.Expr.Type().Kind() != TYPE {
			c.report("Type expected", f.Expr.Loc())
		} else {
			functionScope.returnType = ReadTypeExpression(f.Expr)
		}
	}

	if f.Body != nil {
		c.PushScope(functionScope)
		f.Body.Check(c)
		c.DropScope()
	}
}

func (f FunctionExpression) Loc() tokenizer.Loc {
	loc := tokenizer.Loc{Start: f.Params.Loc().Start, End: tokenizer.Position{}}
	if f.Body == nil {
		loc.End = f.Expr.Loc().End
	} else {
		loc.End = f.Body.Loc().End
	}
	return loc
}
func (f FunctionExpression) Type() ExpressionType {
	// FIXME: return type
	return Function{f.Params.Type(), nil}
}

func ParseFunctionExpression(p *Parser) Expression {
	expr := parseTupleExpression(p)

	tuple, ok := expr.(TupleExpression)
	if !ok {
		return tuple
	}
	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.SLIM_ARR && next.Kind() != tokenizer.FAT_ARR {
		return tuple
	}
	operator := p.tokenizer.Consume()

	next = p.tokenizer.Peek()
	if next.Kind() == tokenizer.LBRACE {
		p.report("Expression expected", next.Loc())
	}

	res := FunctionExpression{tuple, operator, ParseExpression(p), nil}
	if operator.Kind() == tokenizer.FAT_ARR {
		res.Body = ParseBody(p)
	}
	return res
}
