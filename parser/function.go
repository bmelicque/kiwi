package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type FunctionExpression struct {
	params   TupleExpression
	operator tokenizer.Token // -> or =>
	expr     Expression      // return value for '->', return type for '=>'
	body     *Body
}

func checkParam(c *Checker, param Expression) (string, bool) {
	typedExpression, ok := param.(TypedExpression)
	if !ok {
		c.report("Function parameter expected (name: type)", param.Loc())
		return "", false
	}

	tokenExpression, ok := typedExpression.expr.(TokenExpression)
	if !ok {
		c.report("Identifier expected", typedExpression.Loc())
		return "", false
	}

	if tokenExpression.Token.Kind() != tokenizer.IDENTIFIER {
		c.report("Identifier expected", tokenExpression.Loc())
		return "", false
	}

	return tokenExpression.Token.Text(), true
}

func (f FunctionExpression) Check(c *Checker) {
	functionScope := Scope{map[string]*Variable{}, nil, nil}

	for _, param := range f.params.elements {
		name, ok := checkParam(c, param)
		if ok {
			typing := ReadTypeExpression(param.(TypedExpression).typing)
			identifier := param.(TypedExpression).expr.(TokenExpression)
			functionScope.inner[name] = &Variable{identifier.Loc(), typing, []tokenizer.Loc{}, []tokenizer.Loc{}}
		}
	}

	if f.operator.Kind() == tokenizer.SLIM_ARR {
		if f.expr != nil {
			f.expr.Check(c)
			if f.expr.Type(c.scope).Kind() == TYPE {
				c.report("Expression expected", f.expr.Loc())
			}
		}

		if f.body != nil {
			c.report("Function body should be a single statement", f.body.Loc())
		}
		return
	}

	if f.expr != nil {
		f.expr.Check(c)
		if f.expr.Type(c.scope).Kind() != TYPE {
			c.report("Type expected", f.expr.Loc())
		} else {
			functionScope.returnType = ReadTypeExpression(f.expr)
		}
	}

	if f.body != nil {
		c.PushScope(&functionScope)
		f.body.Check(c)
		c.DropScope()
	}
}

func emitParams(e *Emitter, params []Expression) {
	e.Write("(")
	for i, param := range params {
		param.Emit(e)
		if i != len(params)-1 {
			e.Write(", ")
		}
	}
	e.Write(")")
}

func (f FunctionExpression) Emit(e *Emitter) {
	emitParams(e, f.params.elements)

	e.Write(" => ")

	if f.operator.Kind() == tokenizer.SLIM_ARR {
		f.expr.Emit(e)
	} else { // FAT_ARR
		f.body.Emit(e)
	}
}

func (f FunctionExpression) Loc() tokenizer.Loc {
	loc := tokenizer.Loc{Start: f.params.Loc().Start, End: tokenizer.Position{}}
	if f.body == nil {
		loc.End = f.expr.Loc().End
	} else {
		loc.End = f.body.Loc().End
	}
	return loc
}
func (f FunctionExpression) Type(ctx *Scope) ExpressionType {
	// FIXME: return type
	return Function{f.params.Type(ctx).(Tuple), nil}
}

func ParseFunctionExpression(p *Parser) Expression {
	expr := ParseTupleExpression(p)

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
		res.body = ParseBody(p)
	}
	return res
}
