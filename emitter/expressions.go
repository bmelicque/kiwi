package emitter

import (
	"fmt"

	"github.com/bmelicque/test-parser/parser"
)

func (e *Emitter) emitBinaryExpression(expr *parser.BinaryExpression) {
	precedence := Precedence(expr)
	if expr.Left != nil {
		left := Precedence(expr.Left)
		if left < precedence {
			e.write("(")
		}
		e.emit(expr.Left)
		if left < precedence {
			e.write(")")
		}
	}

	e.write(" ")
	e.write(expr.Operator.Text())
	e.write(" ")

	if expr.Right != nil {
		right := Precedence(expr.Right)
		if right < precedence {
			e.write("(")
		}
		e.emit(expr.Right)
		if right < precedence {
			e.write(")")
		}
	}
}

func (e *Emitter) emitBlockExpression(b *parser.Block) {
	if id, ok := e.uninlinables[b]; ok {
		e.write(fmt.Sprintf("_tmp%v", id))
		delete(e.uninlinables, b)
		return
	}

	if len(b.Statements) == 0 {
		e.write("undefined")
		return
	}
	if len(b.Statements) == 1 {
		e.emit(b.Statements[0])
		return
	}
	e.write("(\n")
	e.depth += 1
	for _, statement := range b.Statements {
		e.indent()
		e.emit(statement)
		e.write(",\n")
	}
	e.depth -= 1
	e.indent()
	e.write(")")
}

func (e *Emitter) emitIfExpression(i *parser.IfExpression) {
	e.emit(i.Condition)
	e.write(" ? ")
	e.emitBlockExpression(i.Body)
	e.write(" : ")
	if i.Alternate == nil {
		e.write("undefined")
		return
	}
	switch alternate := i.Alternate.(type) {
	case *parser.Block:
		e.emitBlockExpression(alternate)
	case *parser.IfExpression:
		e.emitIfExpression(alternate)
	}
}

func findMemberByName(members *parser.TupleExpression, name string) parser.Node {
	for _, member := range members.Elements {
		param := member.(*parser.Param)
		text := param.Identifier.Text()
		if text == name {
			return param.Complement
		}
	}
	return nil
}
func (e *Emitter) emitListInstance(constructor *parser.ListTypeExpression, args *parser.TupleExpression) {
	e.write("[")
	c := constructor.Expr
	max := len(args.Elements) - 1
	for _, arg := range args.Elements[:max] {
		e.emitInstance(
			c,
			&parser.TupleExpression{Elements: []parser.Expression{arg}},
		)
		e.write(", ")
	}
	e.emitInstance(
		c,
		&parser.TupleExpression{Elements: []parser.Expression{args.Elements[max]}},
	)
	e.write("]")
}
func (e *Emitter) emitMapInstance(args *parser.TupleExpression) {
	if len(args.Elements) == 0 {
		e.write("new Map()")
		return
	}
	e.write("new Map([")

	max := len(args.Elements) - 1
	for _, arg := range args.Elements[:max] {
		emitMapEntry(e, arg)
		e.write(", ")
	}
	emitMapEntry(e, args.Elements[max])

	e.write("])")
}
func emitMapEntry(e *Emitter, arg parser.Expression) {
	entry := arg.(*parser.Entry)
	var key parser.Expression
	if b, ok := entry.Key.(*parser.BracketedExpression); ok {
		key = b.Expr
	} else {
		key = entry.Key
	}

	e.write("[")
	e.emitExpression(key)
	e.write(", ")
	e.emitExpression(entry.Value)
	e.write("]")
}
func (e *Emitter) emitObjectInstance(constructor *parser.Identifier, args *parser.TupleExpression) {
	e.write("new ")
	e.emit(constructor)
	e.write("(")
	defer e.write(")")
	typing := constructor.Type().(parser.Type).Value.(parser.TypeAlias).Ref.(parser.Object)
	max := len(args.Elements) - 1
	i := 0
	for name := range typing.Members {
		e.emit(findMemberByName(args, name))
		if i != max {
			e.write(", ")
		}
		i++
	}
}
func (e *Emitter) emitSumInstance(constructor *parser.PropertyAccessExpression, args *parser.TupleExpression) {
	e.write("new ")
	e.emit(constructor.Expr)
	e.write("(\"")
	e.emit(constructor.Property)
	e.write("\", ")
	sum := constructor.Expr.(*parser.Identifier).Text()
	cons := constructor.Property.(*parser.Identifier).Text()
	c := e.constructors[sum][cons]
	e.emitInstance(c, args)
	e.write(")")
}
func (e *Emitter) emitInstance(constructor parser.Expression, args *parser.TupleExpression) {
	switch c := constructor.(type) {
	case *parser.ListTypeExpression:
		e.emitListInstance(c, args)
	case *parser.PropertyAccessExpression:
		e.emitSumInstance(c, args)
	case *parser.Identifier:
		if c.Text() == "Map" {
			e.emitMapInstance(args)
		} else {
			e.emitObjectInstance(c, args)
		}
	}
}
func (e *Emitter) emitInstanceExpression(expr *parser.InstanceExpression) {
	e.emitInstance(expr.Typing, expr.Args.Expr.(*parser.TupleExpression))
}

func (e *Emitter) emitCallExpression(expr *parser.CallExpression, await bool) {
	if expr.Callee.Type().(parser.Function).Async && await {
		e.write("await ")
	}
	e.emit(expr.Callee)
	e.write("(")
	defer e.write(")")

	args := expr.Args.Expr.(*parser.TupleExpression).Elements
	max := len(args) - 1
	for i := range args[:max] {
		e.emit(args[i])
		e.write(", ")
	}
	e.emit(args[max])
}

func (e *Emitter) emitComputedAccessExpression(expr *parser.ComputedAccessExpression) {
	if alias, ok := expr.Expr.Type().(parser.TypeAlias); ok && alias.Name == "Map" {
		emitMapElementAccess(e, expr)
		return
	}
	e.emit(expr.Expr)

	if _, ok := expr.Expr.Type().(parser.List); !ok {
		return
	}
	switch prop := expr.Property.Expr.(type) {
	case *parser.RangeExpression:
		e.write(".slice(")
		if prop.Left != nil {
			e.emitExpression(prop.Left)
		} else {
			e.write("0")
		}
		if prop.Right == nil {
			e.write(")")
			return
		}
		e.write(", ")
		e.emitExpression(prop.Right)
		if prop.Operator.Kind() == parser.InclusiveRange {
			e.write("+1")
		}
		e.write(")")
	default:
		e.write("[")
		e.emit(expr.Property.Expr)
		e.write("]")
	}
}
func emitMapElementAccess(e *Emitter, c *parser.ComputedAccessExpression) {
	e.emitExpression(c.Expr)
	e.write(".get(")
	e.emitExpression(c.Property.Expr)
	e.write(")")
}

func (e *Emitter) emitFunctionBody(b *parser.Block, params *parser.TupleExpression) {
	e.write("{")
	if len(b.Statements) == 0 {
		e.write("}")
		return
	}
	e.write("\n")
	e.depth++

	for _, param := range params.Elements {
		if _, ok := param.Type().(parser.Ref); ok {
			continue
		}
		name := param.(*parser.Param).Identifier.Text()
		v, ok := b.Scope().Find(name)
		if !ok {
			panic("variable should be found in current scope...")
		}
		if isMutated(v) {
			e.indent()
			e.write(fmt.Sprintf("%v = structuredClone(%v);\n", name, name))
		}
	}
	max := len(b.Statements) - 1
	for _, statement := range b.Statements[:max] {
		e.indent()
		e.emit(statement)
	}
	e.indent()
	e.write("return ")
	e.emit(b.Statements[max])
	e.write(";\n")
	e.depth--
	e.indent()
	e.write("}\n")
}

func (e *Emitter) emitFunctionExpression(f *parser.FunctionExpression) {
	if f.Type().(parser.Function).Async {
		e.write("async ")
	}
	e.write("(")
	args := f.Params.Expr.(*parser.TupleExpression).Elements
	max := len(args) - 1
	for i := range args[:max] {
		param := args[i].(*parser.Param)
		e.emit(param.Identifier)
		e.write(", ")
	}
	e.emit(args[max].(*parser.Param).Identifier)
	e.write(") => ")

	params := f.Params.Expr.(*parser.TupleExpression)
	e.emitFunctionBody(f.Body, params)
}

func (e *Emitter) emitIdentifier(i *parser.Identifier) {
	text := i.Token.Text()
	if text == e.thisName {
		e.write("this")
		return
	}
	e.write(getSanitizedName(text))
}

func (e *Emitter) emitPropertyAccessExpression(p *parser.PropertyAccessExpression) {
	e.emit(p.Expr)
	if _, ok := p.Expr.Type().(parser.Tuple); ok {
		e.write("[")
		e.emit(p.Property)
		e.write("]")
	} else {
		e.write(".")
		e.emit(p.Property)
	}
}

func (e *Emitter) emitRangeExpression(r *parser.RangeExpression) {
	e.addFlag(RangeFlag)

	e.write("_range(")

	if r.Left != nil {
		e.emit(r.Left)
	} else {
		e.write("0")
	}

	e.write(", ")

	if r.Right != nil {
		e.emit(r.Right)
		if r.Operator.Kind() == parser.InclusiveRange {
			e.write(" + 1")
		}
	} else {
		e.write("1")
	}

	e.write(")")
}

func (e *Emitter) emitTupleExpression(t *parser.TupleExpression) {
	if len(t.Elements) == 1 {
		e.emit(t.Elements[0])
		return
	}
	e.write("[")
	length := len(t.Elements)
	for i, el := range t.Elements {
		e.emit(el)
		if i != length-1 {
			e.write(", ")
		}
	}
	e.write("]")
}

func (e *Emitter) emitUnaryExpression(u *parser.UnaryExpression) {
	switch u.Operator.Kind() {
	case parser.AsyncKeyword:
		e.emitCallExpression(u.Operand.(*parser.CallExpression), false)
	case parser.AwaitKeyword:
		e.write("await ")
		e.emitExpression(u.Operand)
	case parser.Bang:
		e.write("!")
		e.emitExpression(u.Operand)
	case parser.TryKeyword:
		e.emitExpression(u.Operand)
	case parser.BinaryAnd:
		e.write("function (_) { return arguments.length ? void (")
		e.emit(u.Operand)
		e.write(" = _) : ")
		e.emit(u.Operand)
		e.write(" }")
	case parser.Mul:
		e.emit(u.Operand)
		e.write("()")
	}
}
