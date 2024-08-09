package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type ParserError struct {
	Message string
	Loc     tokenizer.Loc
}

func (e ParserError) Error() string {
	return e.Message
}

type Parser struct {
	tokenizer tokenizer.Tokenizer
	errors    []ParserError
}

func (p *Parser) report(message string, loc tokenizer.Loc) {
	p.errors = append(p.errors, ParserError{message, loc})
}
func (p Parser) GetReport() []ParserError {
	return p.errors
}

func MakeParser(tokenizer tokenizer.Tokenizer, scope Scope) *Parser {
	return &Parser{tokenizer, nil}
}

func (p *Parser) ParseProgram() []Node {
	statements := []Node{}

	for p.tokenizer.Peek().Kind() != tokenizer.EOF {
		statements = append(statements, ParseStatement(p))
		next := p.tokenizer.Peek().Kind()
		if next == tokenizer.EOL {
			p.tokenizer.DiscardLineBreaks()
		} else if next != tokenizer.EOF {
			p.report("End of line expected", p.tokenizer.Peek().Loc())
		}
	}

	return statements
}

func ParseStatement(p *Parser) Node {
	switch p.tokenizer.Peek().Kind() {
	case tokenizer.IF_KW:
		return ParseIf(p)
	case tokenizer.FOR_KW:
		return ParseForLoop(p)
	case tokenizer.RETURN_KW:
		return ParseReturn(p)
	default:
		return ParseAssignment(p)
	}
}

func ParseExpression(p *Parser) Node {
	expr := ParseRange(p)
	// TODO: stop at line breaks?
	// TODO: handle EOF
	// TODO: provide a recover token? (e.g. parse until COMMA or EOL for example)
	// for expr == nil {
	// 	expr = BinaryExpression{}.Parse(p)
	// }
	return expr
}
