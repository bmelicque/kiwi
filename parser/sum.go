package parser

import "slices"

type SumTypeConstructor struct {
	Name   *Identifier
	Params *ParenthesizedExpression // contains *TupleExpression
}

func (s SumTypeConstructor) Loc() Loc {
	var start, end Position
	if s.Name != nil {
		start = s.Name.Loc().Start
	} else {
		start = s.Params.Loc().Start
	}
	if s.Params != nil {
		end = s.Params.Loc().End
	} else {
		end = s.Name.Loc().End
	}
	return Loc{start, end}
}

func (s *SumTypeConstructor) typeCheck(p *Parser) {
	if s.Params == nil || s.Params.Expr == nil {
		return
	}
	tuple := s.Params.Expr.(*TupleExpression)
	for i := range tuple.Elements {
		tuple.Elements[i].typeCheck(p)
		if tuple.Elements[i].Type().Kind() != TYPE {
			p.report("Type expected", tuple.Elements[i].Loc())
		}
	}
}

type SumType struct {
	Members []SumTypeConstructor
	typing  ExpressionType
	start   Position
}

func (s *SumType) getChildren() []Node {
	children := []Node{}
	for i := range s.Members {
		if s.Members[i].Name != nil {
			children = append(children, s.Members[i].Name)
		}
		if s.Members[i].Params != nil {
			children = append(children, s.Members[i].Params)
		}
	}
	return children
}

func (s *SumType) Loc() Loc {
	return Loc{
		Start: s.start,
		End:   s.Members[len(s.Members)-1].Loc().End,
	}
}
func (s *SumType) Type() ExpressionType { return s.typing }
func (s *SumType) typeCheck(p *Parser) {
	memberTypes := map[string]Function{}
	for i := range s.Members {
		s.Members[i].typeCheck(p)
		if s.Members[i].Name != nil {
			name := s.Members[i].Name.Text()
			memberTypes[name] = getSumConstructorType(s.Members[i])
		}
	}
	typing := Sum{memberTypes}
	for name := range memberTypes {
		constructor := memberTypes[name]
		constructor.Returned = typing
		memberTypes[name] = constructor
	}
	s.typing = Type{typing}
}

func (p *Parser) parseSumType() Expression {
	if p.Peek().Kind() != BinaryOr {
		return p.parseTaggedExpression()
	}

	start := p.Peek().Loc().Start
	constructors := []SumTypeConstructor{}
	expected := []TokenKind{BinaryOr, EOL, EOF}
	for p.Peek().Kind() == BinaryOr {
		p.Consume()
		constructor := parseSumTypeConstructor(p)
		constructors = append(constructors, constructor)
		if !slices.Contains(expected, p.Peek().Kind()) {
			recover(p, BinaryOr)
		}
		p.DiscardLineBreaks()
	}
	if len(constructors) < 2 {
		p.report("At least 2 constructors expected", constructors[0].Loc())
	}

	return &SumType{Members: constructors, start: start}
}

func parseSumTypeConstructor(p *Parser) SumTypeConstructor {
	identifier := parseSumTypeConstructorName(p)
	var params *ParenthesizedExpression
	if p.Peek().Kind() == LeftParenthesis {
		params = p.parseParenthesizedExpression()
		params.Expr = makeTuple(params.Expr)
	}
	return SumTypeConstructor{identifier, params}
}
func parseSumTypeConstructorName(p *Parser) *Identifier {
	token := p.parseToken()
	if token == nil {
		return nil
	}
	identifier, ok := token.(*Identifier)
	if !ok || !identifier.IsType() {
		p.report("Type identifier expected for type constructor", token.Loc())
		return &Identifier{Token: token.(Token)}
	}
	return identifier
}

func getSumConstructorType(member SumTypeConstructor) Function {
	if member.Params == nil || member.Params.Expr == nil {
		return Function{}
	}

	tuple := member.Params.Expr.(*TupleExpression)
	tu := Tuple{make([]ExpressionType, len(tuple.Elements))}
	for i := range tuple.Elements {
		t, ok := tuple.Elements[i].Type().(Type)
		if ok {
			tu.elements[i] = t.Value
		} else {
			tu.elements[i] = Unknown{}
		}
	}
	return Function{Params: &tu}
}
