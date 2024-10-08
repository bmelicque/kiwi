package parser

type ParserError struct {
	Message string
	Loc     Loc
}

func (e ParserError) Error() string {
	return e.Message
}

type Parser struct {
	Tokenizer
	errors            []ParserError
	multiline         bool
	allowEmptyExpr    bool
	allowBraceParsing bool
	allowCallExpr     bool
}

func (p *Parser) report(message string, loc Loc) {
	p.errors = append(p.errors, ParserError{message, loc})
}
func (p Parser) GetReport() []ParserError {
	return p.errors
}

func MakeParser(tokenizer Tokenizer) *Parser {
	return &Parser{Tokenizer: tokenizer, allowBraceParsing: true, allowCallExpr: true}
}

func (p *Parser) ParseProgram() []Node {
	statements := []Node{}

	for p.Peek().Kind() != EOF {
		statements = append(statements, p.parseStatement())
		next := p.Peek().Kind()
		if next == EOL {
			p.DiscardLineBreaks()
		} else if next != EOF {
			p.report("End of line expected", p.Peek().Loc())
		}
	}

	return statements
}

func (p *Parser) parseStatement() Node {
	switch p.Peek().Kind() {

	case BreakKeyword, ContinueKeyword, ReturnKeyword:
		return p.parseExit()
	default:
		return p.parseAssignment()
	}
}

func ParseExpression(p *Parser) Node {
	switch p.Peek().Kind() {
	case ForKeyword:
		return p.parseForExpression()
	case IfKeyword:
		return p.parseIf()
	case MatchKeyword:
		return p.parseMatchExpression()
	default:
		return p.parseTupleExpression()
	}
}
