package emitter

import (
	"github.com/bmelicque/test-parser/parser"
)

// Check if a variable is mutated. Re-assigns are not accounted for.
// e.g. `object.key = value` is listed, not `variable = value`
func isMutated(v *parser.Variable) bool {
	switch v.Typing.(type) {
	case parser.Boolean, parser.Void, parser.Number, parser.String:
		return false
	}
	writes := v.Writes()
	for _, write := range writes {
		switch write := write.(type) {
		case *parser.Assignment:
			if _, ok := write.Pattern.(*parser.Identifier); ok {
				return true
			}
		case *parser.UnaryExpression:
			return true
		default:
			panic("Invalid type for writes")
		}
	}
	return false
}

func implementsNode(t parser.ExpressionType) bool {
	alias, ok := t.(parser.TypeAlias)
	if !ok {
		return false
	}
	var node parser.Trait
	for _, member := range parser.DomLib().Members {
		if member.Name == "Node" {
			node = member.Type.(parser.Type).Value.(parser.TypeAlias).Ref.(parser.Trait)
			break
		}
	}
	return alias.Implements(node)
}

func isReferenced(identifier *parser.Identifier) bool {
	scope := identifier.GetScope()
	if scope == nil {
		return false
	}
	v := scope.FindLocal(identifier.Text())
	return v != nil && v.HasDirectRef()
}
