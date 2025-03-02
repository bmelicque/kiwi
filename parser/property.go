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
	// TODO: handle <nil>.<nil>
	var start, end Position
	if p.Expr != nil {
		start = p.Expr.Loc().Start
	} else {
		start = p.Property.Loc().Start
	}

	if p.Property != nil {
		end = p.Property.Loc().End
	} else {
		end = p.Expr.Loc().End
	}

	return Loc{start, end}
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

	if p.Peek().Kind() == LeftBrace {
		switch l := left.(type) {
		case *ParenthesizedExpression:
			return parseTraitExpression(p, l)
		case *Identifier:
			return &Param{l, parseTraitExpression(p, nil)}
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
		expr.typing = Invalid{}
		return
	}
	if _, ok := property.Type().(Number); !ok {
		expr.typing = Invalid{}
		return
	}
	number, err := strconv.Atoi(property.Text())
	if err != nil {
		p.error(property, IntegerExpected)
		expr.typing = Invalid{}
		return
	}
	elements := deref(expr.Expr.Type()).(Tuple).Elements
	if number > len(elements)-1 || number < 0 {
		p.error(property, OutOfRange, len(elements), number)
		expr.typing = Invalid{}
		return
	}
	expr.typing = elements[number]
}

// check accessing a sum type's subconstructor: SumType.Constructor
func typeCheckSumConstructorAccess(p *Parser, expr *PropertyAccessExpression) {
	property, ok := expr.Property.(*Identifier)
	if !ok {
		expr.typing = Invalid{}
		return
	}
	name := property.Token.Text()

	expr.typing = getSumTypeConstructor(expr.Expr.Type().(Type), name)
	if expr.typing == (Invalid{}) {
		p.error(expr.Property, PropertyDoesNotExist, name, expr.Expr.Type())
	}
}

func getSumTypeConstructor(t Type, name string) ExpressionType {
	alias, ok := t.Value.(TypeAlias)
	if !ok {
		return Invalid{}
	}

	sum, ok := alias.Ref.(Sum)
	if !ok {
		return Invalid{}
	}

	constructor, ok := sum.Members[name]
	if !ok {
		return Invalid{}
	}

	return Function{Params: &constructor, Returned: alias}
}

// check accessing an object's property or method: object.property
func typeCheckPropertyAccess(p *Parser, expr *PropertyAccessExpression) {
	if reportPrivateFromOtherModule(p, expr) {
		return
	}
	property, ok := expr.Property.(*Identifier)
	if expr.Property != nil && !ok {
		expr.typing = Invalid{}
		return
	}
	var name string
	if property != nil {
		name = property.Token.Text()
	}

	switch t := deref(expr.Expr.Type()).(type) {
	case TypeAlias:
		switch t.Ref.(type) {
		case Trait:
			expr.typing = t.Ref.(Trait).Members[name]
		case Object:
			typing := getCheckedAliasProperty(t, name)
			if len(typing) == 1 {
				expr.typing = typing[0]
			} else if len(typing) > 1 {
				p.error(expr.Property, MultipleEmbeddedProperties, name)
				expr.typing = Invalid{}
			} else {
				p.error(expr.Property, PropertyDoesNotExist, name, expr.Expr.Type())
				expr.typing = Invalid{}
			}
		case Sum:
			expr.typing = t.Methods[name]
		}
	case Module:
		expr.typing, _ = t.GetOwned(name)
	case List:
		expr.typing = getListMethod(t, name)
	}
	if expr.typing == nil {
		p.error(expr.Property, PropertyDoesNotExist, name, expr.Expr.Type())
		expr.typing = Invalid{}
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
	if i.IsPrivate() && alias.From != p.filePath {
		p.error(expr, PrivateProperty, i.Text(), alias.From)
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
	res, _ := object.GetOwned(name)
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
			Returned: makeResultType(Void{}, nil),
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
	if t.Receiver == nil {
		return t.Def.loc
	}
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
	t.buildType()
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
func (t *TraitExpression) buildType() {
	members := t.Def.Type().(Type).Value.(Object).flatten()
	trait := map[string]ExpressionType{}
	for _, member := range members {
		// handle possible duplicates
		_, exists := trait[member.Name]
		if !exists {
			trait[member.Name] = member.Type
		}
	}
	var text string
	if t.Receiver != nil {
		text = t.Receiver.Expr.(*Identifier).Text()
	}
	t.typing = Type{Trait{
		Self:    Generic{Name: text},
		Members: trait,
	}}
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
