package parser

import (
	"fmt"
	"slices"
)

func recover(p *Parser, at TokenKind) bool {
	next := p.Peek()
	start := next.Loc().Start
	end := start
	recovery := []TokenKind{at, EOL, EOF}
	for ; !slices.Contains(recovery, next.Kind()); next = p.Peek() {
		end = p.Consume().Loc().End
	}
	// FIXME: token text
	p.report(fmt.Sprintf("'%v' expected", at), Loc{Start: start, End: end})
	return next.Kind() == at
}

func addTypeParamsToScope(scope *Scope, bracketed *BracketedExpression) {
	tuple := bracketed.Expr.(*TupleExpression)
	for i := range tuple.Elements {
		addTypeParamToScope(scope, tuple.Elements[i].(*Param))
	}
}

func addTypeParamToScope(scope *Scope, param *Param) {
	var constraint ExpressionType
	if param.Complement != nil {
		if t, ok := param.Complement.Type().(Type); ok {
			constraint = t.Value
		}
	}
	name := param.Identifier.Text()
	t := Type{TypeAlias{
		Name: name,
		Ref:  Generic{Name: name, Constraints: constraint},
	}}
	scope.Add(name, param.Loc(), t)
}

// If the given is a result, return its "Ok" type.
// Else return the given type.
func getHappyType(t ExpressionType) ExpressionType {
	if alias, ok := t.(TypeAlias); ok && alias.Name == "!" {
		return alias.Ref.(Sum).getMember("Ok")
	}
	return t
}

// If the given is a result, return its "Err" type.
// Else return nil.
func getErrorType(t ExpressionType) ExpressionType {
	if alias, ok := t.(TypeAlias); ok && alias.Name == "!" {
		return alias.Ref.(Sum).getMember("Err")
	}
	return nil
}

// Check if an expression can be taken as an argument for the ref operator.
// Such an expression can only be identifiers or nested accesses.
func isReferencable(expr Expression) bool {
	if _, ok := expr.(*Identifier); ok {
		return true
	}
	for {
		switch e := expr.(type) {
		case *Identifier:
			return true
		case *InstanceExpression:
			expr = e.Typing
		case *PropertyAccessExpression:
			expr = e.Expr
		case *ComputedAccessExpression:
			expr = e.Expr
		default:
			return false
		}
	}
}
