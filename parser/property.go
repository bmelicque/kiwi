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
		if p.Peek().Kind() == LeftBrace {
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

	switch t := deref(expr.Expr.Type()).(type) {
	case TypeAlias:
		expr.typing = getAliasProperty(t, name)
	case List:
		expr.typing = getListMethod(t, name)
	}
	if expr.typing == nil {
		p.error(expr.Property, PropertyDoesNotExist, name)
		expr.typing = Unknown{}
	}
}

// Get property of given name in aliased object.
// Also checks in alias's methods.
func getAliasProperty(t TypeAlias, name string) ExpressionType {
	if method, ok := t.Methods[name]; ok {
		return method
	}
	object, ok := t.Ref.(Object)
	if !ok {
		return nil
	}
	res, _ := object.get(name)
	return res
}

func getListMethod(l List, name string) ExpressionType {
	switch name {
	case "has":
		return Function{
			Params:   &Tuple{[]ExpressionType{Number{}}},
			Returned: Boolean{},
		}
	case "get":
		return Function{
			Params: &Tuple{[]ExpressionType{Number{}}},
			// FIXME: proper error type
			Returned: makeResultType(l.Element, nil),
		}
	case "set":
		return Function{
			Params: &Tuple{[]ExpressionType{Number{}, l.Element}},
			// FIXME: proper error type
			Returned: makeResultType(Nil{}, nil),
		}
	default:
		return nil
	}
}

type TraitExpression struct {
	Receiver *ParenthesizedExpression // Receiver.Expr is an Identifier
	Def      *BracedExpression        // contains *TupleExpression
}

func (t *TraitExpression) getChildren() []Node {
	return []Node{t.Receiver, t.Def}
}

func (t *TraitExpression) Loc() Loc {
	return Loc{t.Receiver.loc.Start, t.Def.loc.End}
}
func (t *TraitExpression) Type() ExpressionType {
	members := t.Def.Type().(Type).Value.(Object).Members
	trait := map[string]ExpressionType{}
	for _, member := range members {
		trait[member.Name] = member.Type
	}
	return Trait{
		Self:    Generic{Name: t.Receiver.Expr.(*Identifier).Text()},
		Members: trait,
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
	block := p.parseBlock()
	p.allowCallExpr = outer

	braced := getValidatedTraitMethods(p, block)
	return &TraitExpression{
		Receiver: left.(*ParenthesizedExpression),
		Def:      braced,
	}
}

func getValidatedTraitMethods(p *Parser, b *Block) *BracedExpression {
	tuple := &TupleExpression{Elements: make([]Expression, len(b.Statements))}
	for i, s := range b.Statements {
		param, ok := s.(*Param)
		if !ok {
			p.error(s, ParameterExpected)
		}
		tuple.Elements[i] = param
	}
	tuple.reportDuplicatedParams(p)
	return &BracedExpression{tuple, b.loc}
}
