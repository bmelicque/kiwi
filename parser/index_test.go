package parser

import "testing"

func testParserErrors(t *testing.T, p *Parser, expected int) {
	if len(p.errors) == expected {
		return
	}
	t.Logf(
		"Error on test %v. Expected %v error(s), got %v:\n",
		t.Name(),
		expected,
		len(p.errors),
	)
	for _, err := range p.errors {
		t.Log(err.Text())
	}
	t.Fail()
}
