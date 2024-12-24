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
			p.error(u.Operand, PromiseExpected, u.Operand.Type())
		}
	case Bang:
		switch u.Operand.Type().(type) {
		case Type, Boolean:
		default:
			p.error(u.Operand, TypeOrBoolExpected, u.Operand.Type())
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
			p.error(u.Operand, RefExpected, u.Operand.Type())
			return
		}
	case QuestionMark:
		if _, ok := u.Operand.Type().(Type); !ok {
			p.error(u.Operand, TypeExpected)
		}
	case TryKeyword:
		alias, ok := u.Operand.Type().(TypeAlias)
		if !ok || alias.Name != "!" {
			p.error(u.Operand, ResultExpected, u.Operand.Type())
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
		p.error(call, FunctionExpected)
		return
	}
	if !f.Async {
		p.error(u, UnneededAsync)
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
		p.error(l, TypeExpected)
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
			p.error(expr, NotReferenceable)
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
		p.error(operand, CallExpressionExpected)
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
	switch p.Peek().Kind() {
	case LeftParenthesis:
		return p.parseFunctionExpression(brackets)
	case LeftBrace:
		return parseAnonymousList(p, brackets)
	default:
		expr := parseInnerUnary(p)
		if brackets != nil && brackets.Expr != nil {
			p.error(brackets, UnexpectedExpression)
		}
		return &ListTypeExpression{brackets, expr}
	}
}

func parseAnonymousList(p *Parser, brackets *BracketedExpression) *InstanceExpression {
	args := p.parseBracedExpression()
	args.Expr = makeTuple(args.Expr)
	return &InstanceExpression{
		Typing: &ListTypeExpression{brackets, nil},
		Args:   args,
	}
}
