package checker

import (
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

type For struct {
	Keyword   tokenizer.Token
	Condition Expression
	Body      Block
	typing    ExpressionType
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

func (f For) Type() ExpressionType { return f.typing }

type RangeDeclaration struct {
	Pattern  Expression
	Range    RangeExpression
	Constant bool
}

type ForRange struct {
	Keyword     tokenizer.Token
	Declaration RangeDeclaration
	Body        Block
	typing      ExpressionType
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
func (f ForRange) Type() ExpressionType { return f.typing }

func (c *Checker) checkLoop(node parser.ForExpression) Expression {
	switch node.Statement.(type) {
	case parser.Assignment:
		l := checkForRangeLoop(c, node)
		l.typing = checkLoopType(c, l)
		return l
	default:
		l := checkForLoop(c, node)
		l.typing = checkLoopType(c, l)
		return l
	}
}

func checkForLoop(c *Checker, node parser.ForExpression) For {
	expr, _ := c.Check(node.Statement).(Expression)
	if expr == nil || expr.Type().Kind() != BOOLEAN {
		c.report("Expected boolean condition or range declaration", node.Statement.Loc())
	}

	c.pushScope(NewScope(LoopScope))
	defer c.dropScope()
	body := c.checkBlock(*node.Body)
	return For{
		Keyword:   node.Keyword,
		Condition: expr,
		Body:      body,
	}
}
func checkForRangeLoop(c *Checker, node parser.ForExpression) ForRange {
	declaration, ok := c.Check(node.Statement).(VariableDeclaration)
	if !ok {
		c.report("Expected boolean condition or range declaration", node.Statement.Loc())
	}

	r, ok := declaration.Initializer.(RangeExpression)
	if !ok {
		c.report("Expected range type", declaration.Initializer.Loc())
	}

	c.pushScope(NewScope(LoopScope))
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

	body := c.checkBlock(*node.Body)
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

func checkLoopType(c *Checker, expr Expression) ExpressionType {
	var body Block
	switch expr := expr.(type) {
	case For:
		body = expr.Body
	case ForRange:
		body = expr.Body
	}
	breaks := []Exit{}
	findBreakStatements(body, &breaks)
	if len(breaks) == 0 {
		return Primitive{NIL}
	}
	var t ExpressionType
	if breaks[0].Value != nil {
		t = breaks[0].Value.Type()
	} else {
		t = Primitive{NIL}
	}
	for _, b := range breaks[1:] {
		if t == (Primitive{NIL}) && b.Value != nil {
			c.report("No value expected", b.Value.Loc())
		}
		if t != (Primitive{NIL}) && !t.Extends(b.Value.Type()) {
			c.report("Type doesn't match the type inferred from first break", b.Value.Loc())
		}
	}
	return t
}

func findBreakStatements(node Node, results *[]Exit) {
	if node == nil {
		return
	}
	if n, ok := node.(Exit); ok {
		if n.Operator.Kind() == tokenizer.BREAK_KW {
			*results = append(*results, n)
		}
		return
	}
	switch node := node.(type) {
	case Block:
		for _, statement := range node.Statements {
			findBreakStatements(statement, results)
		}
	case If:
		findBreakStatements(node.Block, results)
		findBreakStatements(node.Alternate, results)
	}
}
