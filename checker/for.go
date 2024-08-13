package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type For struct {
	Keyword   tokenizer.Token
	Condition Expression
	Body      Body
}

func (f For) Loc() tokenizer.Loc {
	loc := f.Keyword.Loc()
	if f.Body.loc != (tokenizer.Loc{}) {
		loc.End = f.Body.Loc().End
	} else if f.Condition != nil {
		loc.End = f.Condition.Loc().End
	}
	return loc
}

type RangeDeclaration struct {
	Pattern  Expression
	Range    RangeExpression
	Constant bool
}

type ForRange struct {
	Keyword     tokenizer.Token
	Declaration RangeDeclaration
	Body        Body
}

func (f ForRange) Loc() tokenizer.Loc {
	loc := f.Keyword.Loc()
	if f.Body.loc != (tokenizer.Loc{}) {
		loc.End = f.Body.Loc().End
	} else if f.Declaration != (RangeDeclaration{}) {
		loc.End = f.Declaration.Range.Loc().End
	}
	return loc
}

func (c *Checker) checkForLoop(node parser.For) For {
	expr, _ := c.Check(node.Statement).(Expression)
	if expr == nil || expr.Type().Kind() != BOOLEAN {
		c.report("Expected boolean condition or range declaration", node.Statement.Loc())
	}

	body := c.checkBody(*node.Body)
	return For{
		Keyword:   node.Keyword,
		Condition: expr,
		Body:      body,
	}
}
func (c *Checker) checkForRangeLoop(node parser.For) ForRange {
	declaration, ok := c.Check(node).(VariableDeclaration)
	if !ok {
		c.report("Expected boolean condition or range declaration", node.Statement.Loc())
	}

	r, ok := declaration.Initializer.(RangeExpression)
	if !ok {
		c.report("Expected range type", declaration.Initializer.Loc())
	}

	c.pushScope(NewShadowScope())
	defer c.dropScope()
	switch pattern := declaration.Pattern.(type) {
	case Identifier:
		c.scope.Add(pattern.Text(), pattern.Loc(), r.Type().(Range).operands)
	case TupleExpression:
		// TODO: FIXME:
		c.report("Expected identifier", declaration.Pattern.Loc())
	default:
		c.report("Expected identifier", declaration.Pattern.Loc())
	}

	body := c.checkBody(*node.Body)
	return ForRange{
		Keyword: node.Keyword,
		Declaration: RangeDeclaration{
			Pattern:  declaration.Pattern,
			Range:    r,
			Constant: declaration.Constant,
		},
		Body: body,
	}
}

func (c *Checker) checkLoop(node parser.For) Node {
	switch node.Statement.(type) {
	case VariableDeclaration:
		return c.checkForRangeLoop(node)
	default:
		return c.checkForLoop(node)
	}
}
