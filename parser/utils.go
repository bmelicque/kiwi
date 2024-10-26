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

func (p *Parser) addTypeArgsToScope(args *TupleExpression, params []Generic) {
	var l int
	if args != nil {
		l = len(args.Elements)
	}

	if l > len(params) {
		loc := args.Elements[len(params)].Loc()
		loc.End = args.Elements[len(args.Elements)-1].Loc().End
		p.report("Too many type arguments", loc)
	}

	for i, param := range params {
		var loc Loc
		var t ExpressionType
		if i < l {
			arg := args.Elements[i]
			loc = arg.Loc()
			typing, ok := arg.Type().(Type)
			if ok {
				t = typing.Value
			} else {
				p.report("Typing expected", arg.Loc())
			}
		}
		if t != nil && param.Value != nil && !param.Value.Extends(t) {
			p.report("Type doesn't match", args.Elements[i].Loc())
		} else {
			params[i].Value = t
		}
		p.scope.Add(param.Name, loc, Type{Generic{Name: param.Name, Value: t}})
		v, _ := p.scope.Find(param.Name)
		v.readAt(loc)
	}
}

func addTypeParamsToScope(scope *Scope, params Params) {
	for _, param := range params.Params {
		if param.Complement == nil {
			name := param.Identifier.Text()
			t := Type{TypeAlias{Name: name, Ref: Generic{Name: name}}}
			scope.Add(name, param.Loc(), t)
		}
	}
}

// If the given is a result, return its "Ok" type.
// Else return the given type.
func getHappyType(t ExpressionType) ExpressionType {
	if alias, ok := t.(TypeAlias); ok && alias.Name == "Result" {
		return alias.Ref.(Sum).getMember("Ok")
	}
	return t
}

// If the given is a result, return its "Err" type.
// Else return nil.
func getErrorType(t ExpressionType) ExpressionType {
	if alias, ok := t.(TypeAlias); ok && alias.Name == "Result" {
		return alias.Ref.(Sum).getMember("Err")
	}
	return nil
}
