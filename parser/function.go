package parser

import (
	"fmt"
	"reflect"

	"github.com/bmelicque/test-parser/tokenizer"
)

type FunctionExpression struct {
	TypeParams *BracketedExpression
	Params     *ParenthesizedExpression
	Operator   tokenizer.Token // -> or =>
	Expr       Node            // return value for '->', return type for '=>'
	Body       *Body
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

func (p *Parser) parseFunctionExpression() Node {
	var brackets *BracketedExpression
	var paren *ParenthesizedExpression

	if p.tokenizer.Peek().Kind() == tokenizer.LBRACKET {
		expr := p.parseListTypeExpression()
		if list, ok := getListType(expr); ok {
			return list
		}
		brackets, paren = getBracketedParenthesized(expr)
	}

	if paren == nil && p.tokenizer.Peek().Kind() == tokenizer.LPAREN {
		pa := p.parseParenthesizedExpression()
		paren = &pa
	}

	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.SLIM_ARR && next.Kind() != tokenizer.FAT_ARR {
		if brackets == nil {
			return *paren
		}
		if paren == nil {
			return *brackets
		}
		return ListTypeExpression{*brackets, *paren}
	}
	operator := p.tokenizer.Consume()

	next = p.tokenizer.Peek()
	if next.Kind() == tokenizer.LBRACE {
		p.report("Expression expected", next.Loc())
	}

	var expr Node
	if operator.Kind() == tokenizer.FAT_ARR {
		old := p.allowBraceParsing
		p.allowBraceParsing = false
		expr = ParseRange(p)
		p.allowBraceParsing = old
	} else {
		expr = ParseRange(p)
	}
	res := FunctionExpression{brackets, paren, operator, expr, nil}
	if operator.Kind() == tokenizer.FAT_ARR {
		res.Body = p.parseBody()
	}
	return res
}

func getListType(node Node) (ListTypeExpression, bool) {
	list, ok := node.(ListTypeExpression)
	if !ok {
		return ListTypeExpression{}, false
	}

	if _, ok = list.Type.(ParenthesizedExpression); ok {
		return ListTypeExpression{}, false
	}

	return list, true
}

func getBracketedParenthesized(node Node) (*BracketedExpression, *ParenthesizedExpression) {
	if brackets, ok := node.(BracketedExpression); ok {
		return &brackets, nil
	}

	list, ok := node.(ListTypeExpression)
	if !ok {
		panic(fmt.Sprintf("parseArrayType should've returned a BracketedExpression or an ArrayType, got %#v\n", reflect.TypeOf(node)))
	}

	paren, ok := list.Type.(ParenthesizedExpression)
	if !ok {
		panic(fmt.Sprintf("expected Parenthesized expression after getArrayType, got %#v\n", reflect.TypeOf(node)))
	}
	return &list.Bracketed, &paren
}
