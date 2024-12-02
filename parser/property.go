package parser

import (
	"strconv"
)

// Expr.Property
type PropertyAccessExpression struct {
	Expr     Expression
	Property Expression
	typing   ExpressionType
}

func (p *PropertyAccessExpression) getChildren() []Node {
	children := []Node{p.Expr}
	if p.Property != nil {
		children = append(children, p.Property)
	}
	return children
}

func (p *PropertyAccessExpression) Loc() Loc {
	return Loc{
		Start: p.Expr.Loc().Start,
		End:   p.Property.Loc().End,
	}
}
func (p *PropertyAccessExpression) Type() ExpressionType { return p.typing }

func (expr *PropertyAccessExpression) typeCheck(p *Parser) {
	expr.Expr.typeCheck(p)
	switch deref(expr.Expr.Type()).(type) {
	case Tuple:
		typeCheckTupleIndexAccess(p, expr)
	case Type:
		typeCheckSumConstructorAccess(p, expr)
	default:
		typeCheckPropertyAccess(p, expr)
	}
}

func parsePropertyAccess(p *Parser, left Expression) Expression {
	p.Consume() // .
	if _, ok := left.(*ParenthesizedExpression); ok {
		if p.Peek().Kind() == LeftParenthesis {
			return parseTraitExpression(p, left)
		}
	}
	prop := fallback(p)
	switch prop.(type) {
	case *Identifier, *Literal:
	default:
		p.error(prop, IdentifierExpected)
	}
	return &PropertyAccessExpression{Expr: left, Property: prop}
}

// check accessing a tuple's index: tuple.0
func typeCheckTupleIndexAccess(p *Parser, expr *PropertyAccessExpression) {
	property, ok := expr.Property.(*Literal)
	if !ok {
		expr.typing = Unknown{}
		return
	}
	if _, ok := property.Type().(Number); !ok {
		expr.typing = Unknown{}
		return
	}
	number, err := strconv.Atoi(property.Text())
	if err != nil {
		p.error(property, IntegerExpected)
		expr.typing = Unknown{}
		return
	}
	elements := deref(expr.Expr.Type()).(Tuple).Elements
	if number > len(elements)-1 || number < 0 {
		p.error(property, OutOfRange, len(elements), number)
		expr.typing = Unknown{}
		return
	}
	expr.typing = elements[number]
}

// check accessing a sum type's subconstructor: SumType.Constructor
func typeCheckSumConstructorAccess(p *Parser, expr *PropertyAccessExpression) {
	property, ok := expr.Property.(*Identifier)
	if !ok {
		expr.typing = Unknown{}
		return
	}
	name := property.Token.Text()

	expr.typing = getSumTypeConstructor(expr.Expr.Type().(Type), name)
	if expr.typing == (Unknown{}) {
		p.error(expr.Property, PropertyDoesNotExist, name)
	}
}

func getSumTypeConstructor(t Type, name string) ExpressionType {
	alias, ok := t.Value.(TypeAlias)
	if !ok {
		return Unknown{}
	}

	sum, ok := alias.Ref.(Sum)
	if !ok {
		return Unknown{}
	}

	constructor, ok := sum.Members[name]
	if !ok {
		return Unknown{}
	}

	return constructor
}

// check accessing an object's property or method: object.property
func typeCheckPropertyAccess(p *Parser, expr *PropertyAccessExpression) {
	property, ok := expr.Property.(*Identifier)
	if expr.Property != nil && !ok {
		expr.typing = Unknown{}
		return
	}
	var name string
	if property != nil {
		name = property.Token.Text()
	}

	alias, ok := deref(expr.Expr.Type()).(TypeAlias)
	if !ok {
		p.error(expr.Property, PropertyDoesNotExist, name)
		expr.typing = Unknown{}
		return
	}
	if method, ok := alias.Methods[name]; ok {
		expr.typing = method
		return
	}

	object, ok := alias.Ref.(Object)
	if !ok {
		p.error(expr.Property, PropertyDoesNotExist, name)
		expr.typing = Unknown{}
		return
	}

	t, ok := object.Members[name]
	if !ok {
		p.error(expr.Property, PropertyDoesNotExist, name)
		expr.typing = Unknown{}
		return
	}
	expr.typing = t
}

type TraitExpression struct {
	Receiver *ParenthesizedExpression // Receiver.Expr is an Identifier
	Def      *ParenthesizedExpression // contains *TupleExpression
}

func (t *TraitExpression) getChildren() []Node {
	return []Node{t.Receiver, t.Def}
}

func (t *TraitExpression) Loc() Loc {
	return Loc{t.Receiver.loc.Start, t.Def.loc.End}
}
func (t *TraitExpression) Type() ExpressionType {
	return Trait{
		Self:    Generic{Name: t.Receiver.Expr.(*Identifier).Text()},
		Members: t.Def.Type().(Type).Value.(Object).Members,
	}
}
func (t *TraitExpression) typeCheck(p *Parser) {
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()

	receiver := t.Receiver.Expr.(*Identifier)
	if receiver != nil {
		p.scope.Add(
			receiver.Text(),
			receiver.Loc(),
			Generic{Name: receiver.Text()},
		)
	}

	for _, element := range t.Def.Expr.(*TupleExpression).Elements {
		param, ok := element.(*Param)
		if !ok {
			continue
		}
		typing, ok := param.Complement.Type().(Type)
		if !ok {
			p.error(param.Complement, FunctionTypeExpected)
			continue
		}
		if _, ok := typing.Value.(Function); !ok {
			p.error(param.Complement, FunctionTypeExpected)
		}
	}
}

func parseTraitExpression(p *Parser, left Expression) Expression {
	outer := p.allowCallExpr
	p.allowCallExpr = false
	paren := p.parseParenthesizedExpression()
	p.allowCallExpr = outer

	if paren != nil {
		paren.Expr = makeTuple(paren.Expr)
		validateFunctionParams(p, paren)
	}
	return &TraitExpression{
		Receiver: left.(*ParenthesizedExpression),
		Def:      paren,
	}
}
