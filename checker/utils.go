package checker

import "github.com/bmelicque/test-parser/parser"

func addTypeParamsToScope(scope *Scope, params Params) {
	for _, param := range params.Params {
		if param.Complement == nil {
			name := param.Identifier.Text()
			t := Type{TypeAlias{Name: name, Ref: Generic{Name: name}}}
			scope.Add(name, param.Loc(), t)
		} else {
			// TODO: constrained generic
		}
	}
}

func checkTypeIdentifier(c *Checker, node parser.Node) (Identifier, bool) {
	token, ok := node.(parser.TokenExpression)
	if !ok {
		return Identifier{}, false
	}

	identifier, ok := c.checkToken(token, false).(Identifier)
	if !ok {
		return Identifier{}, false
	}
	return identifier, identifier.isType
}
