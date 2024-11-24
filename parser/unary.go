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
	if u.Operand == nil {
		return
	}
	u.Operand.typeCheck(p)
	switch u.Operator.Kind() {
	case AsyncKeyword:
		typeCheckAsyncExpression(p, u)
	case AwaitKeyword:
		alias, ok := u.Operand.Type().(TypeAlias)
		if !ok || alias.Name != "..." {
			p.report("Promise expected", u.Operand.Loc())
		}
	case Bang:
		switch u.Operand.Type().(type) {
		case Type, Boolean:
		default:
			p.report("Type or boolean expected with '!' operator", u.Operand.Loc())
		}
	case BinaryAnd:
		identifier := getReferencedIdentifier(u.Operand)
		if identifier == nil {
			return
		}
		v, ok := p.scope.Find(identifier.Text())
		if ok {
			v.writeAt(u)
		}
	case Mul:
		if _, ok := u.Operand.Type().(Ref); !ok {
			p.report("Reference expected", u.Operand.Loc())
			return
		}
	case QuestionMark:
		if _, ok := u.Operand.Type().(Type); !ok {
			p.report("Type expected with question mark operator", u.Operand.Loc())
		}
	case TryKeyword:
		alias, ok := u.Operand.Type().(TypeAlias)
		if !ok || alias.Name != "!" {
			p.report("Result type expected", u.Operand.Loc())
		}
	default:
		panic(fmt.Sprintf("Operator '%v' not implemented!", u.Operator.Kind()))
	}
}
func typeCheckAsyncExpression(p *Parser, u *UnaryExpression) {
	call, ok := u.Operand.(*CallExpression)
	if !ok || call.Callee == nil {
		return
	}
	f, ok := call.Callee.Type().(Function)
	if !ok {
		p.report("Function expected", call.Loc())
		return
	}
	if !f.Async {
		p.report("'async' keyword has no effect in this expression", u.Loc())
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
	if u.Operand == nil {
		return Unknown{}
	}
	switch u.Operator.Kind() {
	case AsyncKeyword:
		return makePromise(u.Operand.Type())
	case AwaitKeyword:
		alias, ok := u.Operand.Type().(TypeAlias)
		if !ok || alias.Name != "..." {
			return Unknown{}
		}
		t, _ := alias.Params[0].Value.build(nil, nil)
		return t
	case Bang:
		t := u.Operand.Type()
		if ty, ok := t.(Type); ok {
			t = ty.Value
			return Type{makeResultType(t, nil)}
		} else {
			return Boolean{}
		}
	case BinaryAnd:
		switch t := u.Operand.Type().(type) {
		case Type:
			return Type{Ref{t.Value}}
		default:
			return Ref{t}
		}
	case Mul:
		ref, ok := u.Operand.Type().(Ref)
		if !ok {
			return Unknown{}
		}
		return ref.To
	case QuestionMark:
		t := u.Operand.Type()
		if ty, ok := t.(Type); ok {
			t = ty.Value
		}
		return Type{makeOptionType(t)}
	case TryKeyword:
		alias, ok := u.Operand.Type().(TypeAlias)
		if !ok || alias.Name != "!" {
			return Unknown{}
		}
		return alias.Ref.(Sum).getMember("Ok")
	default:
		return Unknown{}
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
	if l.Expr == nil {
		return
	}
	l.Expr.typeCheck(p)
	if _, ok := l.Expr.Type().(Type); !ok {
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
		return Type{List{Unknown{}}}
	}
	return Type{List{t.Value}}
}

func (p *Parser) parseUnaryExpression() Expression {
	switch p.Peek().Kind() {
	case AsyncKeyword, AwaitKeyword, Bang, BinaryAnd, Mul, QuestionMark, TryKeyword:
		token := p.Consume()
		expr := parseInnerUnary(p)
		if token.Kind() == AsyncKeyword {
			validateAsyncOperand(p, expr)
		}
		if token.Kind() == BinaryAnd && !isReferencable(expr) {
			p.report("Cannot reference such an expression", expr.Loc())
		}
		return &UnaryExpression{token, expr}
	case LeftBracket:
		return parseListTypeExpression(p)
	default:
		return p.parseAccessExpression()
	}
}

func validateAsyncOperand(p *Parser, operand Expression) {
	if operand == nil {
		return
	}
	if _, ok := operand.(*CallExpression); !ok {
		p.report("Call expression expected", operand.Loc())
	}
}

func parseInnerUnary(p *Parser) Expression {
	memBrace := p.allowBraceParsing
	p.allowBraceParsing = false
	expr := p.parseUnaryExpression()
	p.allowBraceParsing = memBrace
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
