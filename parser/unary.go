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
		checkAsyncExpression(p, u)
	case AwaitKeyword:
		checkAwaitExpression(p, u)
	case Bang:
		checkBangExpression(p, u)
	case BinaryAnd:
		checkReferenceExpression(p, u)
	case Mul:
		checkDerefExpression(p, u)
	case QuestionMark:
		checkOptionType(p, u)
	case TryKeyword:
		checkTryExpression(p, u)
	default:
		panic(fmt.Sprintf("Operator '%v' not implemented!", u.Operator.Kind()))
	}
}
func checkAsyncExpression(p *Parser, u *UnaryExpression) {
	call, ok := u.Operand.(*CallExpression)
	if !ok || call.Callee == nil {
		return
	}
	f, ok := call.Callee.Type().(Function)
	if ok && !f.Async {
		p.error(u, UnneededAsync)
	}
}
func checkAwaitExpression(p *Parser, u *UnaryExpression) {
	alias, ok := u.Operand.Type().(TypeAlias)
	if !ok || alias.Name != "..." {
		p.error(u.Operand, PromiseExpected, u.Operand.Type())
	}
}

// error type | boolean neg
func checkBangExpression(p *Parser, u *UnaryExpression) {
	switch u.Operand.Type().(type) {
	case Type, Boolean:
	default:
		p.error(u.Operand, TypeOrBoolExpected, u.Operand.Type())
	}
}
func checkReferenceExpression(p *Parser, u *UnaryExpression) {
	identifier := getReferencedIdentifier(u.Operand)
	if identifier == nil {
		return
	}
	v, ok := p.scope.Find(identifier.Text())
	if !ok {
		return
	}
	v.readAt(u.Loc())
	v.writeAt(u)
	if _, ok := u.Operand.(*Identifier); ok {
		v.hasDirectRef = true
	}
}
func checkDerefExpression(p *Parser, u *UnaryExpression) {
	if _, ok := u.Operand.Type().(Ref); !ok {
		p.error(u.Operand, RefExpected, u.Operand.Type())
	}
}
func checkOptionType(p *Parser, u *UnaryExpression) {
	if _, ok := u.Operand.Type().(Type); !ok {
		p.error(u.Operand, TypeExpected)
	}
}
func checkTryExpression(p *Parser, u *UnaryExpression) {
	alias, ok := u.Operand.Type().(TypeAlias)
	if !ok || alias.Name != "!" {
		p.error(u.Operand, ResultExpected, u.Operand.Type())
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
		return Invalid{}
	}
	switch u.Operator.Kind() {
	case AsyncKeyword:
		return getAsyncType(u)
	case AwaitKeyword:
		return getAwaitedType(u)
	case Bang:
		return getBangType(u)
	case BinaryAnd:
		return getRefType(u)
	case Mul:
		return getDerefType(u)
	case QuestionMark:
		return getOptionType(u)
	case TryKeyword:
		return getTryType(u)
	default:
		return Invalid{}
	}
}
func getAsyncType(u *UnaryExpression) ExpressionType {
	var t ExpressionType
	if c, ok := u.Operand.(*CallExpression); ok {
		t = c.Type()
	} else {
		t = Invalid{}
	}
	return makePromise(t)
}
func getAwaitedType(u *UnaryExpression) ExpressionType {
	raw := u.Operand.Type()
	alias, ok := raw.(TypeAlias)
	if !ok || alias.Name != "..." {
		return raw
	}
	t, _ := alias.Params[0].Value.build(nil, nil)
	return t
}
func getBangType(u *UnaryExpression) ExpressionType {
	switch t := u.Operand.Type().(type) {
	case Type:
		return Type{makeResultType(t.Value, nil)}
	case Boolean:
		return Boolean{}
	default:
		return Invalid{}
	}
}
func getRefType(u *UnaryExpression) ExpressionType {
	switch t := u.Operand.Type().(type) {
	case Type:
		return Type{Ref{t.Value}}
	default:
		return Ref{t}
	}
}
func getDerefType(u *UnaryExpression) ExpressionType {
	ref, ok := u.Operand.Type().(Ref)
	if !ok {
		return Invalid{}
	}
	return ref.To
}
func getOptionType(u *UnaryExpression) ExpressionType {
	var t ExpressionType
	if ty, ok := u.Operand.Type().(Type); ok {
		t = ty.Value
	} else {
		t = Invalid{}
	}
	return Type{makeOptionType(t)}
}
func getTryType(u *UnaryExpression) ExpressionType {
	alias, ok := u.Operand.Type().(TypeAlias)
	if !ok || alias.Name != "!" {
		return Invalid{}
	}
	return alias.Ref.(Sum).getMember("Ok")
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
	if l.Expr == nil {
		return Type{List{Invalid{}}}
	}
	t, ok := l.Expr.Type().(Type)
	if !ok {
		return Type{List{Invalid{}}}
	}
	return Type{List{t.Value}}
}

func (p *Parser) parseUnaryExpression() Expression {
	switch p.Peek().Kind() {
	case AsyncKeyword, AwaitKeyword, Bang, BinaryAnd, Mul, QuestionMark, TryKeyword:
		token := p.Consume()
		if token.Kind() == QuestionMark && p.Peek().Kind() == LeftBrace {
			return parseInferredInstance(p, &UnaryExpression{token, nil})
		}
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
		return parseInferredInstance(p, &ListTypeExpression{brackets, nil})
	default:
		expr := parseInnerUnary(p)
		if brackets != nil && brackets.Expr != nil {
			p.error(brackets, UnexpectedExpression)
		}
		return &ListTypeExpression{brackets, expr}
	}
}
