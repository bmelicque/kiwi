package parser

type SumTypeConstructor struct {
	Name   *Identifier
	Params *Params
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

type SumType struct {
	Members []SumTypeConstructor
	typing  ExpressionType
	start   Position
}

func (s SumType) Loc() Loc {
	return Loc{
		Start: s.start,
		End:   s.Members[len(s.Members)-1].Loc().End,
	}
}
func (s SumType) Type() ExpressionType { return s.typing }

func (p *Parser) parseSumType() Expression {
	if p.Peek().Kind() != BinaryOr {
		return p.parseTypedExpression()
	}

	start := p.Peek().Loc().Start
	memberTypes := map[string]*Function{}
	constructors := []SumTypeConstructor{}
	for p.Peek().Kind() == BinaryOr {
		p.Consume()
		constructor := parseSumTypeConstructor(p)
		constructors = append(constructors, constructor)
		if constructor.Name != nil {
			name := constructor.Name.Text()
			memberTypes[name] = getSumConstructorType(constructor)
		}
		handleSumTypeBadTokens(p)
		p.DiscardLineBreaks()
	}
	if len(constructors) < 2 {
		p.report("At least 2 constructors expected", constructors[0].Loc())
	}
	typing := Sum{memberTypes}
	for name := range memberTypes {
		memberTypes[name].Returned = &typing
	}
	return SumType{Members: constructors, start: start, typing: Type{typing}}
}

func handleSumTypeBadTokens(p *Parser) {
	err := false
	var start, end Position
	for p.Peek().Kind() != EOL && p.Peek().Kind() != EOF && p.Peek().Kind() != BinaryOr {
		token := p.Consume()
		if !err {
			err = true
			start = token.Loc().Start
		}
		end = token.Loc().End
	}
	if err {
		p.report("EOL or '|' expected", Loc{Start: start, End: end})
	}
}

func parseSumTypeConstructor(p *Parser) SumTypeConstructor {
	identifier := parseSumTypeConstructorName(p)
	var params *Params
	if p.Peek().Kind() == LeftParenthesis {
		params = p.getValidatedTypeList(*p.parseParenthesizedExpression())
	}
	return SumTypeConstructor{identifier, params}
}
func parseSumTypeConstructorName(p *Parser) *Identifier {
	token := p.parseToken(true)
	if token == nil {
		return nil
	}
	identifier, ok := token.(*Identifier)
	if !ok || !identifier.isType {
		p.report("Type identifier expected for type constructor", token.Loc())
		return &Identifier{Token: token.(Token), isType: true}
	}
	return identifier
}

func getSumConstructorType(member SumTypeConstructor) *Function {
	if member.Params == nil {
		return &Function{}
	}

	tuple := Tuple{make([]ExpressionType, len(member.Params.Params))}
	for i, param := range member.Params.Params {
		t, ok := param.Type().(Type)
		if ok {
			tuple.elements[i] = t.Value
		} else {
			tuple.elements[i] = Primitive{UNKNOWN}
		}
	}
	return &Function{Params: &tuple}
}
