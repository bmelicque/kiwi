package parser

import (
	"slices"

	"github.com/bmelicque/test-parser/tokenizer"
)

// TODO: random access expression

type CallExpression struct {
	Callee   Node
	TypeArgs *BracketedExpression
	Args     *ParenthesizedExpression
}

func (c CallExpression) Loc() tokenizer.Loc {
	loc := tokenizer.Loc{}
	if c.Callee != nil {
		loc.Start = c.Callee.Loc().Start
	} else {
		loc.Start = c.Args.Loc().Start
	}
	if c.Args != nil {
		loc.End = c.Args.Loc().End
	} else {
		loc.End = c.Callee.Loc().End
	}
	return loc
}

type PropertyAccessExpression struct {
	Expr     Node
	Property Node
	method   bool
}

func (p PropertyAccessExpression) IsMethod() bool { return p.method }

func (p PropertyAccessExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: p.Expr.Loc().Start,
		End:   p.Property.Loc().End,
	}
}

// TODO: InstanceExpression
type ObjectExpression struct {
	Typing   Node
	TypeArgs *BracketedExpression
	Members  []Node
	loc      tokenizer.Loc
}

func (o ObjectExpression) Loc() tokenizer.Loc { return o.loc }

var operators = []tokenizer.TokenKind{tokenizer.LBRACKET, tokenizer.LPAREN, tokenizer.DOT, tokenizer.LBRACE}

func (p *Parser) parseAccessExpression() Node {
	expression := fallback(p)
	next := p.tokenizer.Peek()
	for slices.Contains(operators, next.Kind()) {
		switch next.Kind() {
		case tokenizer.LBRACKET:
			var typeArgs BracketedExpression
			if next.Kind() == tokenizer.LBRACKET {
				typeArgs = p.parseBracketedExpression()
			}
			next = p.tokenizer.Peek()
			if next.Kind() == tokenizer.LPAREN {
				args := p.parseParenthesizedExpression()
				expression = CallExpression{expression, nil, &args}
			} else if next.Kind() == tokenizer.LBRACE && p.allowBraceParsing {
				p.tokenizer.Consume()
				var members []Node
				ParseList(p, tokenizer.RBRACE, func() {
					members = append(members, p.parseTypedExpression())
				})
				loc := tokenizer.Loc{Start: expression.Loc().Start}
				if p.tokenizer.Peek().Kind() != tokenizer.RBRACE {
					p.report("'}' expected", p.tokenizer.Peek().Loc())
				} else {
					loc.End = p.tokenizer.Consume().Loc().End
				}
				expression = ObjectExpression{
					Typing:  expression,
					Members: members,
					loc:     loc,
				}
			} else {
				expression = CallExpression{expression, &typeArgs, nil}
			}
		case tokenizer.LPAREN:
			args := p.parseParenthesizedExpression()
			expression = CallExpression{expression, nil, &args}
		case tokenizer.DOT:
			p.tokenizer.Consume()
			property := fallback(p)
			expression = PropertyAccessExpression{
				Expr:     expression,
				Property: property,
			}

		case tokenizer.LBRACE:
			if !p.allowBraceParsing {
				return expression
			}
			// TODO: parseTuple
			p.tokenizer.Consume()
			var members []Node
			ParseList(p, tokenizer.RBRACE, func() {
				members = append(members, p.parseTypedExpression())
			})
			loc := tokenizer.Loc{Start: expression.Loc().Start}
			if p.tokenizer.Peek().Kind() != tokenizer.RBRACE {
				p.report("'}' expected", p.tokenizer.Peek().Loc())
			} else {
				loc.End = p.tokenizer.Consume().Loc().End
			}
			expression = ObjectExpression{
				Typing:  expression,
				Members: members,
				loc:     loc,
			}
		}
		next = p.tokenizer.Peek()
	}
	return expression
}

func fallback(p *Parser) Node {
	switch p.tokenizer.Peek().Kind() {
	case tokenizer.LBRACKET, tokenizer.LPAREN:
		return p.parseFunctionExpression()
	// case tokenizer.LBRACKET:
	// 	return ListExpression{}.Parse(p)
	case tokenizer.LBRACE:
		if p.allowBraceParsing {
			return p.parseObjectDefinition()
		}
	}
	return p.parseTokenExpression()
}
