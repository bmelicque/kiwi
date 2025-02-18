package parser

import "io"

type Parser struct {
	tokenizer
	errors            []ParserError
	scope             *Scope
	filePath          string
	writing           Node
	multiline         bool
	allowEmptyExpr    bool
	allowBraceParsing bool
	allowCallExpr     bool
	preventColon      bool // don't parse expressions like 'identifier: value'

	// If true, declarations are considered as being part of an if statement.
	// For example: 'if s Some := option {}'.
	// Some patterns are allowed in if statements, but not in regular declarations.
	conditionalDeclaration bool
}

func (p *Parser) error(node Node, kind ErrorKind, comp ...interface{}) {
	var complements [2]interface{}
	switch len(comp) {
	case 0:
	case 1:
		complements[0] = comp[0]
	case 2:
		complements[0] = comp[0]
		complements[1] = comp[1]
	default:
		panic("too many complements")
	}
	err := ParserError{
		Node:        node,
		Kind:        kind,
		Complements: complements,
	}
	p.errors = append(p.errors, err)
}

func MakeParser(reader io.Reader) *Parser {
	tokenizer := NewTokenizer(reader)
	scope := NewScope(ProgramScope)
	scope.outer = &std
	return &Parser{
		tokenizer:         *tokenizer,
		scope:             scope,
		allowBraceParsing: true,
		allowCallExpr:     true,
	}
}

func (p *Parser) pushScope(scope *Scope) {
	scope.outer = p.scope
	p.scope = scope
}

func (p *Parser) dropScope() {
	for _, info := range p.scope.variables {
		if len(info.reads) == 0 {
			p.error(&Block{loc: info.declaredAt}, UnusedVariable)
		}
	}
	p.scope = p.scope.outer
}

func (p *Parser) parseStatement() Node {
	switch p.Peek().Kind() {
	case BreakKeyword, ContinueKeyword, ReturnKeyword, ThrowKeyword:
		return p.parseExit()
	case UseKeyword:
		return p.parseUseDirective()
	default:
		return p.parseAssignment()
	}
}

func (p *Parser) parseExpression() Expression {
	switch p.Peek().Kind() {
	case ForKeyword:
		return p.parseForExpression()
	case IfKeyword:
		return p.parseIfExpression()
	case MatchKeyword:
		return p.parseMatchExpression()
	default:
		return p.parseTupleExpression()
	}
}
