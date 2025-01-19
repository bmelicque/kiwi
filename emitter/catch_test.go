package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestEmitCatchStatement(t *testing.T) {
	type test struct {
		name               string
		node               *parser.CatchExpression
		expectedStatement  string
		expectedExpression string
	}
	tests := []test{
		{
			name: "with identifier",
			// result catch _ { 0 }
			node: &parser.CatchExpression{
				Left:       &parser.Identifier{Token: testToken{kind: parser.Name, value: "result"}},
				Identifier: &parser.Identifier{Token: testToken{kind: parser.Name, value: "_"}},
				Body: parser.MakeBlock([]parser.Node{
					&parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "0"}},
				}),
			},
			expectedStatement:  "try {\n    result;\n} catch (_) {\n    0;\n}\n",
			expectedExpression: "let __tmp0;\ntry {\n    __tmp0 = result;\n} catch (_) {\n    __tmp0 = 0;\n}\n__tmp0",
		},
		{
			name: "without identifier",
			// result catch { 0 }
			node: &parser.CatchExpression{
				Left:       &parser.Identifier{Token: testToken{kind: parser.Name, value: "result"}},
				Identifier: nil,
				Body: parser.MakeBlock([]parser.Node{
					&parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "0"}},
				}),
			},
			expectedStatement:  "try {\n    result;\n} catch (_) {\n    0;\n}\n",
			expectedExpression: "let __tmp0;\ntry {\n    __tmp0 = result;\n} catch (_) {\n    __tmp0 = 0;\n}\n__tmp0",
		},
		{
			name: "with exiting statement",
			// result catch { return }
			node: &parser.CatchExpression{
				Left:       &parser.Identifier{Token: testToken{kind: parser.Name, value: "result"}},
				Identifier: nil,
				Body: parser.MakeBlock([]parser.Node{
					&parser.Exit{Operator: testToken{kind: parser.ReturnKeyword}},
				}),
			},
			expectedStatement:  "try {\n    result;\n} catch (_) {\n    return;\n}\n",
			expectedExpression: "let __tmp0;\ntry {\n    __tmp0 = result;\n} catch (_) {\n    return;\n}\n__tmp0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emitter := makeEmitter()
			emitter.emit(tt.node)
			text := emitter.string()

			if text != tt.expectedStatement {
				t.Errorf("Expected statement:\n%v\ngot:\n%v", tt.expectedStatement, text)
			}
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emitter := makeEmitter()
			emitter.extractUninlinables(tt.node)
			emitter.emitExpression(tt.node)
			text := emitter.string()

			if text != tt.expectedExpression {
				t.Fatalf("Expected expression:\n%v\ngot:\n%v", tt.expectedExpression, text)
			}
		})
	}
}
