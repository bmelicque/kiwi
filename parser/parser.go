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
	tokenizer         tokenizer.Tokenizer
	errors            []ParserError
	multiline         bool
	allowEmptyExpr    bool
	allowBraceParsing bool
	allowCallExpr     bool
}

func (p *Parser) report(message string, loc tokenizer.Loc) {
	p.errors = append(p.errors, ParserError{message, loc})
}
func (p Parser) GetReport() []ParserError {
	return p.errors
}

func MakeParser(tokenizer tokenizer.Tokenizer) *Parser {
	return &Parser{tokenizer: tokenizer, allowBraceParsing: true, allowCallExpr: true}
}

func (p *Parser) ParseProgram() []Node {
	statements := []Node{}

	for p.tokenizer.Peek().Kind() != tokenizer.EOF {
		statements = append(statements, p.parseStatement())
		next := p.tokenizer.Peek().Kind()
		if next == tokenizer.EOL {
			p.tokenizer.DiscardLineBreaks()
		} else if next != tokenizer.EOF {
			p.report("End of line expected", p.tokenizer.Peek().Loc())
		}
	}

	return statements
}

func (p *Parser) parseStatement() Node {
	switch p.tokenizer.Peek().Kind() {
	case tokenizer.FOR_KW:
		return ParseForLoop(p)
	case tokenizer.BREAK_KW, tokenizer.CONTINUE_KW, tokenizer.RETURN_KW:
		return p.parseExit()
	default:
		return p.parseAssignment()
	}
}

func ParseExpression(p *Parser) Node {
	switch p.tokenizer.Peek().Kind() {
	case tokenizer.IF_KW:
		return p.parseIf()
	case tokenizer.MATCH_KW:
		return p.parseMatchExpression()
	default:
		return p.parseTupleExpression()
	}
}
