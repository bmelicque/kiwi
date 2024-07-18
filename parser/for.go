package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type For struct {
	keyword   tokenizer.Token
	Statement Statement // ExpressionStatement holding a condition OR Assignment
	Body      *Body
}

func (f For) Loc() tokenizer.Loc {
	loc := f.keyword.Loc()
	if f.Body != nil {
		loc.End = f.Body.Loc().End
	} else if f.Statement != nil {
		loc.End = f.Statement.Loc().End
	}
	return loc
}

func (f For) Check(c *Checker) {
	if f.Statement == nil {
		return
	}

	scope := Scope{map[string]*Variable{}, nil, nil}
	scope.returnType = c.scope.returnType
	switch statement := f.Statement.(type) {
	case ExpressionStatement:
		if statement.Expr.Type(c.scope) != (Primitive{BOOLEAN}) {
			c.report("Boolean expected", statement.Loc())
		}
	case Assignment:
		declared, ok := statement.Declared.(TokenExpression)
		if !ok || declared.Token.Kind() != tokenizer.IDENTIFIER {
			c.report("Identifier expected", statement.Declared.Loc())
		} else {
			text := declared.Token.Text()
			if text != "_" {
				scope.inner[declared.Token.Text()] = &Variable{
					declaredAt: declared.Loc(),
					typing:     statement.Initializer.Type(c.scope),
				}
			}
		}
		if statement.Operator.Kind() != tokenizer.DECLARE && statement.Operator.Kind() != tokenizer.DEFINE {
			c.report("':=' or '::' expected", statement.Operator.Loc())
		}
		if _, ok := statement.Initializer.(RangeExpression); !ok {
			c.report("Range expression expected", statement.Initializer.Loc())
		}
	default:
		c.report("Condition or declaration expected", statement.Loc())
	}

	c.PushScope(&scope)
	f.Body.Check(c)
	c.DropScope()
}

func ParseForLoop(p *Parser) Statement {
	statement := For{}
	statement.keyword = p.tokenizer.Consume()

	if p.tokenizer.Peek().Kind() != tokenizer.LBRACE {
		statement.Statement = ParseAssignment(p)
	}

	statement.Body = ParseBody(p)

	return statement
}
