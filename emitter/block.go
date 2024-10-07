package emitter

import "github.com/bmelicque/test-parser/checker"

type hoistedBlock struct {
	block           *checker.Block
	label           string
	parentStatement *checker.Node
}

type blockHoister struct {
	blocks []hoistedBlock
}

func (b blockHoister) findStatementBlocks(statement *checker.Node) []hoistedBlock {
	blocks := []hoistedBlock{}
	for _, h := range b.blocks {
		if h.parentStatement == statement {
			blocks = append(blocks, h)
		}
	}
	return blocks
}

func (b blockHoister) findBlockLabel(block *checker.Block) (string, bool) {
	for _, h := range b.blocks {
		if h.block == block {
			return h.label, true
		}
	}
	return "", false
}

// doesn't look for nested blocks
func findBlocks(node checker.Node, blocks *[]checker.Block) {
	if node == nil {
		return
	}
	switch node := node.(type) {
	case checker.Assignment:
		findBlocks(node.Value, blocks)
	case checker.BinaryExpression:
		findBlocks(node.Left, blocks)
		findBlocks(node.Right, blocks)
	case checker.Block:
		*blocks = append(*blocks, node)
	case checker.CallExpression:
		findBlocks(node.Callee, blocks)
		for _, arg := range node.Args.Params {
			findBlocks(arg, blocks)
		}
	case checker.ComputedAccessExpression:
		findBlocks(node.Expr, blocks)
		findBlocks(node.Property, blocks)
	case checker.Exit:
		findBlocks(node.Value, blocks)
	case checker.ExpressionStatement:
		findBlocks(node.Expr, blocks)
	case checker.If:
		*blocks = append(*blocks, node.Block)
		findBlocks(node.Alternate, blocks)
	case checker.MatchExpression:
		for _, c := range node.Cases {
			for _, s := range c.Statements {
				findBlocks(s, blocks)
			}
		}
	case checker.ParenthesizedExpression:
		findBlocks(node.Expr, blocks)
	case checker.PropertyAccessExpression:
		findBlocks(node.Expr, blocks)
	case checker.RangeExpression:
		findBlocks(node.Left, blocks)
		findBlocks(node.Right, blocks)
	case checker.TupleExpression:
		for _, e := range node.Elements {
			findBlocks(e, blocks)
		}
	case checker.VariableDeclaration:
		findBlocks(node.Initializer, blocks)
	}
}

// Check if a statement triggers a block hoisting.
// Sub-blocks are added to the hoisted list
// Return true if the statement needs hoisting
func findHoisted(statement checker.Node, hoisted *[]hoistedBlock) bool {
	var trig bool
	switch statement.(type) {
	case
		checker.Exit,
		checker.For,
		checker.ForRange,
		checker.MethodDeclaration,
		checker.VariableDeclaration:
		trig = true
	}

	blocks := []checker.Block{}
	findBlocks(statement, &blocks)
	for _, block := range blocks {
		var ok bool
		for _, s := range block.Statements {
			ok = ok || findHoisted(s, hoisted)
		}
		if ok {
			el := hoistedBlock{
				block:           &block,
				label:           "", // TODO: id_generator
				parentStatement: &statement,
			}
			*hoisted = append(*hoisted, el)
			trig = trig || ok
		}
	}
	return trig
}
