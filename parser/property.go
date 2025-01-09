package parser

import (
	"slices"
	"strconv"
)

// Expr.Property
type PropertyAccessExpression struct {
	Expr     Expression
	Property Expression
	typing   ExpressionType
}

// consists only of embedded property accesses (a.b.c etc.)
func (p *PropertyAccessExpression) isSimple() bool {
	switch expr := p.Expr.(type) {
	case *Identifier:
		return true
	case *PropertyAccessExpression:
		return expr.isSimple()
	default:
		return false
	}
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
	if l, ok := left.(*ParenthesizedExpression); ok {
		if p.Peek().Kind() == LeftBrace {
			return parseTraitExpression(p, l)
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
	if reportPrivateFromOtherModule(p, expr) {
		return
	}
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
		typing := getCheckedAliasProperty(t, name)
		if len(typing) == 1 {
			expr.typing = typing[0]
		} else if len(typing) > 1 {
			p.error(expr.Property, MultipleEmbeddedProperties, name)
			expr.typing = Unknown{}
		} else {
			p.error(expr.Property, PropertyDoesNotExist, name)
			expr.typing = Unknown{}
		}
	case Module:
		expr.typing, _ = t.getOwned(name)
	case List:
		expr.typing = getListMethod(t, name)
	}
	if expr.typing == nil {
		p.error(expr.Property, PropertyDoesNotExist, name)
		expr.typing = Unknown{}
	}
}
func reportPrivateFromOtherModule(p *Parser, expr *PropertyAccessExpression) bool {
	i, ok := expr.Property.(*Identifier)
	if !ok {
		return false
	}
	t := expr.Expr.Type()
	if a, ok := t.(Type); ok {
		t = a.Value
	}
	alias, ok := t.(TypeAlias)
	if !ok {
		return false
	}
	if i.IsPrivate() && alias.from != p.filePath {
		p.error(expr, PrivateProperty, i.Text(), alias.from)
		return true
	}
	return false
}

func getCheckedAliasProperty(t TypeAlias, name string) []ExpressionType {
	owned := getAliasOwnedProperty(t, name)
	if owned != nil {
		return []ExpressionType{owned}
	}
	o, ok := t.Ref.(Object)
	if !ok {
		return nil
	}
	shallow := o.Embedded
	for {
		types, deep := findShallowlyEmbedded(shallow, name)
		if len(types) > 0 {
			return types
		}
		if len(deep) == 0 {
			return nil
		}
		shallow = deep
	}
}

// returns [found types, deeper layer of member]
func findShallowlyEmbedded(members []ObjectMember, name string) ([]ExpressionType, []ObjectMember) {
	var found bool
	types := []ExpressionType{}
	deep := []ObjectMember{}
	for _, embedded := range members {
		if embedded.Name == name {
			types = append(types, embedded.Type)
			found = true
			continue
		}
		if found {
			continue
		}
		alias, ok := embedded.Type.(TypeAlias)
		if !ok {
			continue
		}
		o, ok := alias.Ref.(Object)
		if !ok {
			continue
		}
		deep = append(deep, o.Embedded...)
	}
	return types, deep
}

// Get property of given name in aliased object.
// Also checks in alias's methods.
func getAliasOwnedProperty(t TypeAlias, name string) ExpressionType {
	if method, ok := t.Methods[name]; ok {
		return method
	}
	object, ok := t.Ref.(Object)
	if !ok {
		return nil
	}
	res, _ := object.getOwned(name)
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
	typing   ExpressionType
}

func (t *TraitExpression) getChildren() []Node {
	return []Node{t.Receiver, t.Def}
}

func (t *TraitExpression) Loc() Loc {
	return Loc{t.Receiver.loc.Start, t.Def.loc.End}
}
func (t *TraitExpression) Type() ExpressionType { return t.typing }
func (t *TraitExpression) typeCheck(p *Parser) {
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()

	typeCheckTraitReceiver(p, t.Receiver)

	// parsing makes sure all elements inside braces{} is a valid,
	// either a type identifier, or a &Param{identifier, function type}
	members := t.Def.Type().(Type).Value.(Object).flatten()
	duplicates := findMemberDuplicates(members)
	for _, duplicate := range duplicates {
		p.error(t, DuplicateIdentifier, duplicate)
	}
	trait := map[string]ExpressionType{}
	for _, member := range members {
		if !slices.Contains(duplicates, member.Name) {
			trait[member.Name] = member.Type
		}
	}
	t.typing = Trait{
		Self:    Generic{Name: t.Receiver.Expr.(*Identifier).Text()},
		Members: trait,
	}
}
func typeCheckTraitReceiver(p *Parser, receiver *ParenthesizedExpression) {
	// receiver is nil in case of .{} shorthand
	if receiver == nil {
		return
	}
	identifier := receiver.Expr.(*Identifier)
	if identifier == nil {
		return
	}
	p.scope.Add(
		identifier.Text(),
		identifier.Loc(),
		Generic{Name: identifier.Text()},
	)
}

func parseTraitExpression(p *Parser, left *ParenthesizedExpression) Expression {
	outer := p.allowCallExpr
	p.allowCallExpr = false
	block := p.parseBlock()
	p.allowCallExpr = outer

	braced := getValidatedTraitMethods(p, block)
	return &TraitExpression{
		Receiver: left,
		Def:      braced,
	}
}

func getValidatedTraitMethods(p *Parser, b *Block) *BracedExpression {
	tuple := &TupleExpression{Elements: make([]Expression, len(b.Statements))}
	i := 0
	for _, s := range b.Statements {
		if expr := getValidatedTraitMethod(p, s); expr != nil {
			tuple.Elements[i] = expr
			i++
		}
	}
	tuple.Elements = tuple.Elements[:i]
	tuple.reportDuplicatedParams(p)
	return &BracedExpression{tuple, b.loc}
}

// returns nil if given Node is not valid
func getValidatedTraitMethod(p *Parser, n Node) Expression {
	switch expr := n.(type) {
	case *Identifier:
		if !expr.IsType() {
			p.error(expr, TypeExpected)
			return nil
		}
		return expr
	case *PropertyAccessExpression:
		return getValidatedEmbedding(p, expr)
	case *Param:
		_, okComplement := expr.Complement.(*FunctionTypeExpression)
		if !okComplement {
			p.error(expr.Complement, FunctionTypeExpected)
		}
		okIdentifier := !expr.Identifier.IsType()
		if !okIdentifier {
			p.error(expr.Identifier, ValueIdentifierExpected)
		}
		if !okComplement || !okIdentifier {
			return nil
		}
		return expr
	default:
		p.error(n, ParameterExpected)
		return nil
	}
}
