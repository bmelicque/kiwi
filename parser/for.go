package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type For struct {
	keyword   tokenizer.Token
	statement Statement // ExpressionStatement holding a condition OR Assignment
	body      *Body
}

func (f For) Loc() tokenizer.Loc {
	loc := f.keyword.Loc()
	if f.body != nil {
		loc.End = f.body.Loc().End
	} else if f.statement != nil {
		loc.End = f.statement.Loc().End
	}
	return loc
}

func (f For) Emit(e *Emitter) {
	if assignment, ok := f.statement.(Assignment); ok {
		e.Write("for (const ")
		assignment.declared.Emit(e)
		e.Write(" of ")
		assignment.initializer.Emit(e)
	} else {
		e.Write("while (")
		if f.statement != nil {
			// FIXME: ';' at the end of the statement, body should handle where to put ';'
			f.statement.Emit(e)
		} else {
			e.Write("true")
		}
	}
	e.Write(") ")

	f.body.Emit(e)
}

func (f For) Check(c *Checker) {
	if f.statement == nil {
		return
	}

	scope := Scope{map[string]*Variable{}, nil, nil}
	scope.returnType = c.scope.returnType
	switch statement := f.statement.(type) {
	case ExpressionStatement:
		if statement.expr.Type(c.scope) != (Primitive{BOOLEAN}) {
			c.report("Boolean expected", statement.Loc())
		}
	case Assignment:
		declared, ok := statement.declared.(TokenExpression)
		if !ok || declared.Token.Kind() != tokenizer.IDENTIFIER {
			c.report("Identifier expected", statement.declared.Loc())
		} else {
			text := declared.Token.Text()
			if text != "_" {
				scope.inner[declared.Token.Text()] = &Variable{
					declaredAt: declared.Loc(),
					typing:     statement.initializer.Type(c.scope),
				}
			}
		}
		if statement.operator.Kind() != tokenizer.DECLARE && statement.operator.Kind() != tokenizer.DEFINE {
			c.report("':=' or '::' expected", statement.operator.Loc())
		}
		if _, ok := statement.initializer.(RangeExpression); !ok {
			c.report("Range expression expected", statement.initializer.Loc())
		}
	default:
		c.report("Condition or declaration expected", statement.Loc())
	}

	c.PushScope(&scope)
	f.body.Check(c)
	c.DropScope()
}

func ParseForLoop(p *Parser) Statement {
	statement := For{}
	statement.keyword = p.tokenizer.Consume()

	if p.tokenizer.Peek().Kind() != tokenizer.LBRACE {
		statement.statement = ParseAssignment(p)
	}

	statement.body = ParseBody(p)

	return statement
}
