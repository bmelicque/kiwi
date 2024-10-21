package emitter

import "github.com/bmelicque/test-parser/parser"

type hoistedBlock struct {
	block           *parser.Block
	label           string
	parentStatement *parser.Node
}

type blockHoister struct {
	blocks []hoistedBlock
}

func (b blockHoister) findStatementBlocks(statement *parser.Node) []hoistedBlock {
	blocks := []hoistedBlock{}
	for _, h := range b.blocks {
		if h.parentStatement == statement {
			blocks = append(blocks, h)
		}
	}
	return blocks
}

func (b blockHoister) findBlockLabel(block *parser.Block) (string, bool) {
	for _, h := range b.blocks {
		if h.block == block {
			return h.label, true
		}
	}
	return "", false
}

// doesn't look for nested blocks
func findBlocks(node parser.Node, blocks *[]*parser.Block) {
	if node == nil {
		return
	}
	switch node := node.(type) {
	case *parser.Assignment:
		findBlocks(node.Value, blocks)
	case *parser.BinaryExpression:
		findBlocks(node.Left, blocks)
		findBlocks(node.Right, blocks)
	case *parser.Block:
		*blocks = append(*blocks, node)
	case *parser.CallExpression:
		findBlocks(node.Callee, blocks)
		for _, arg := range node.Args.Expr.(*parser.TupleExpression).Elements {
			findBlocks(arg, blocks)
		}
	case *parser.ComputedAccessExpression:
		findBlocks(node.Expr, blocks)
		findBlocks(node.Property.Expr, blocks)
	case *parser.Exit:
		findBlocks(node.Value, blocks)
	case *parser.IfExpression:
		*blocks = append(*blocks, node.Body)
		findBlocks(node.Alternate, blocks)
	case *parser.MatchExpression:
		for _, c := range node.Cases {
			for _, s := range c.Statements {
				findBlocks(s, blocks)
			}
		}
	case *parser.ParenthesizedExpression:
		findBlocks(node.Expr, blocks)
	case *parser.PropertyAccessExpression:
		findBlocks(node.Expr, blocks)
	case *parser.RangeExpression:
		findBlocks(node.Left, blocks)
		findBlocks(node.Right, blocks)
	case *parser.TupleExpression:
		for _, e := range node.Elements {
			findBlocks(e, blocks)
		}
	}
}

// Check if a statement triggers a block hoisting.
// Sub-blocks are added to the hoisted list
// Return true if the statement needs hoisting
func findHoisted(statement parser.Node, hoisted *[]hoistedBlock) bool {
	var trig bool
	switch s := statement.(type) {
	case
		*parser.Exit,
		*parser.ForExpression:
		trig = true
	case *parser.Assignment:
		o := s.Operator.Kind()
		trig = o == parser.Declare || o == parser.Define
	}

	blocks := []*parser.Block{}
	findBlocks(statement, &blocks)
	for _, block := range blocks {
		var ok bool
		for _, s := range block.Statements {
			ok = ok || findHoisted(s, hoisted)
		}
		if ok {
			el := hoistedBlock{
				block:           block,
				label:           "", // TODO: id_generator
				parentStatement: &statement,
			}
			*hoisted = append(*hoisted, el)
			trig = trig || ok
		}
	}
	return trig
}
