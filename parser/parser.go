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
	scope             *Scope
	multiline         bool
	allowEmptyExpr    bool
	allowBraceParsing bool
	allowCallExpr     bool

	// If true, declarations are considered as being part of an if statement.
	// For example: 'if Some(s) := option {}'.
	// Some patterns are allowed in if statements, but not in regular declarations.
	conditionalDeclaration bool
}

func (p *Parser) report(message string, loc Loc) {
	p.errors = append(p.errors, ParserError{message, loc})
}
func (p Parser) GetReport() []ParserError {
	return p.errors
}

func MakeParser(tokenizer Tokenizer) *Parser {
	return &Parser{
		Tokenizer:         tokenizer,
		scope:             &std,
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
			p.report("Unused variable", info.declaredAt)
		}
	}
	p.scope = p.scope.outer
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

	for i := range statements {
		statements[i].typeCheck(p)
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

func (p *Parser) parseExpression() Expression {
	return ParseExpression(p)
}

func ParseExpression(p *Parser) Expression {
	switch p.Peek().Kind() {
	case ForKeyword:
		return p.parseForExpression()
	case IfKeyword:
		return p.parseIfExpression()
	case MatchKeyword:
		return p.parseMatchExpression()
	case TryKeyword:
		return p.parseTryExpression()
	default:
		return p.parseTupleExpression()
	}
}
