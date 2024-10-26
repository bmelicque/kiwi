package parser

import "fmt"

type UnaryExpression struct {
	Operator Token
	Operand  Expression
}

func (u *UnaryExpression) getChildren() []Node {
	if u.Operand == nil {
		return []Node{}
	}
	return []Node{u.Operand}
}

func (u *UnaryExpression) typeCheck(p *Parser) {
	u.Operand.typeCheck(p)
	switch u.Operator.Kind() {
	case Bang:
		t := u.Operand.Type().Kind()
		if t != TYPE && t != BOOLEAN {
			p.report("Type or boolean expected wiyh '!' operator", u.Operand.Loc())
		}
	case QuestionMark:
		if u.Operand.Type().Kind() != TYPE {
			p.report("Type expected with question mark operator", u.Operand.Loc())
		}
	default:
		panic(fmt.Sprintf("Operator '%v' not implemented!", u.Operator.Kind()))
	}
}

func (u *UnaryExpression) Loc() Loc {
	loc := u.Operator.Loc()
	if u.Operand != nil {
		loc.End = u.Operand.Loc().End
	}
	return loc
}

func (u *UnaryExpression) Type() ExpressionType {
	switch u.Operator.Kind() {
	case Bang:
		t := u.Operand.Type()
		if ty, ok := t.(Type); ok {
			t = ty.Value
			return Type{makeResultType(t, nil)}
		} else {
			return Primitive{BOOLEAN}
		}
	case QuestionMark:
		t := u.Operand.Type()
		if ty, ok := t.(Type); ok {
			t = ty.Value
		}
		return Type{makeOptionType(t)}
	default:
		return Primitive{UNKNOWN}
	}
}

type ListTypeExpression struct {
	Bracketed *BracketedExpression
	Expr      Expression // Cannot be nil
}

func (l *ListTypeExpression) getChildren() []Node {
	if l.Expr == nil {
		return []Node{}
	}
	return []Node{l.Expr}
}

func (l *ListTypeExpression) typeCheck(p *Parser) {
	l.Expr.typeCheck(p)
	if l.Expr != nil && l.Expr.Type().Kind() != TYPE {
		p.report("Type expected", l.Loc())
	}
}

func (l *ListTypeExpression) Loc() Loc {
	loc := l.Bracketed.Loc()
	if l.Expr != nil {
		loc.End = l.Expr.Loc().End
	}
	return loc
}

func (l *ListTypeExpression) Type() ExpressionType {
	t, ok := l.Expr.Type().(Type)
	if !ok {
		return Type{List{Primitive{UNKNOWN}}}
	}
	return Type{List{t.Value}}
}

func (p *Parser) parseUnaryExpression() Expression {
	switch p.Peek().Kind() {
	case Bang, QuestionMark:
		token := p.Consume()
		expr := parseInnerUnary(p)
		return &UnaryExpression{token, expr}
	case LeftBracket:
		return parseListTypeExpression(p)
	default:
		return p.parseAccessExpression()
	}
}

func parseInnerUnary(p *Parser) Expression {
	memBrace := p.allowBraceParsing
	memCall := p.allowCallExpr
	p.allowBraceParsing = false
	p.allowCallExpr = false
	expr := p.parseUnaryExpression()
	p.allowBraceParsing = memBrace
	p.allowCallExpr = memCall
	return expr
}

func parseListTypeExpression(p *Parser) Expression {
	brackets := p.parseBracketedExpression()
	if p.Peek().Kind() == LeftParenthesis {
		return p.parseFunctionExpression(brackets)
	}
	expr := parseInnerUnary(p)
	if brackets != nil && brackets.Expr != nil {
		p.report("No expression expected for list type", brackets.Loc())
	}
	return &ListTypeExpression{brackets, expr}
}
