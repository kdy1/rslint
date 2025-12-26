package prefer_literal_enum_member

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferLiteralEnumMemberRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferLiteralEnumMemberRule,
		[]rule_tester.ValidTestCase{
			{Code: `enum ValidRegex { A = /test/, }`},
			{Code: `enum ValidString { A = 'test', }`},
			{Code: "enum ValidLiteral { A = `test`, }"},
			{Code: `enum ValidNumber { A = 42, }`},
			{Code: `enum ValidNumber { A = -42, }`},
			{Code: `enum ValidNumber { A = +42, }`},
			{Code: `enum ValidNull { A = null, }`},
			{Code: `enum ValidPlain { A, }`},
			{Code: `enum ValidQuotedKey { 'a', }`},
			{Code: `enum ValidQuotedKeyWithAssignment { 'a' = 1, }`},
			{
				Code:    `enum Foo { A = 1 << 0, B = 1 >> 0, C = 1 >>> 0, D = 1 | 0, E = 1 & 0, F = 1 ^ 0, G = ~1, }`,
				Options: map[string]interface{}{"allowBitwiseExpressions": true},
			},
			{
				Code:    `enum Foo { A = 1 << 0, B = 1 >> 0, C = A | B, }`,
				Options: map[string]interface{}{"allowBitwiseExpressions": true},
			},
			{
				Code:    `enum Foo { A = 1 << 0, B = 1 >> 0, C = Foo.A | Foo.B, }`,
				Options: map[string]interface{}{"allowBitwiseExpressions": true},
			},
			{
				Code:    "enum Foo { A = 1 << 0, B = 1 >> 0, C = Foo['A'] | B, }",
				Options: map[string]interface{}{"allowBitwiseExpressions": true},
			},
			{
				Code:    `enum Foo { A = 1 << 0, B = 1 << 1, C = 1 << 2, D = A | B | C, }`,
				Options: map[string]interface{}{"allowBitwiseExpressions": true},
			},
			{
				Code:    `enum Foo { A = 1 << 0, B = 1 << 1, C = 1 << 2, D = Foo.A | Foo.B | Foo.C, }`,
				Options: map[string]interface{}{"allowBitwiseExpressions": true},
			},
			{
				Code:    `enum Foo { A = 1 << 0, B = 1 << 1, C = 1 << 2, D = Foo.A | (Foo.B & ~Foo.C), }`,
				Options: map[string]interface{}{"allowBitwiseExpressions": true},
			},
			{
				Code:    `enum Foo { A = 1 << 0, B = 1 << 1, C = 1 << 2, D = Foo.A | -Foo.B, }`,
				Options: map[string]interface{}{"allowBitwiseExpressions": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `enum InvalidObject { A = {}, }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 22},
				},
			},
			{
				Code: `enum InvalidArray { A = [], }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 21},
				},
			},
			{
				Code: "enum InvalidTemplateLiteral { A = `foo ${0}`, }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 32},
				},
			},
			{
				Code: `enum InvalidConstructor { A = new Set(), }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 27},
				},
			},
			{
				Code: `enum InvalidExpression { A = 2 + 2, }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 26},
				},
			},
			{
				Code: `enum InvalidExpression { A = delete 2, B = -a, C = void 2, D = ~2, E = !0, }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 26},
					{MessageId: "notLiteral", Line: 1, Column: 41},
					{MessageId: "notLiteral", Line: 1, Column: 49},
					{MessageId: "notLiteral", Line: 1, Column: 61},
					{MessageId: "notLiteral", Line: 1, Column: 69},
				},
			},
			{
				Code: `const variable = 'Test'; enum InvalidVariable { A = 'TestStr', B = 2, C, V = variable, }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 75},
				},
			},
			{
				Code: `enum InvalidEnumMember { A = 'TestStr', B = A, }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 42},
				},
			},
			{
				Code: `const Valid = { A: 2 }; enum InvalidObjectMember { A = 'TestStr', B = Valid.A, }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 68},
				},
			},
			{
				Code: `enum Valid { A, } enum InvalidEnumMember { A = 'TestStr', B = Valid.A, }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 60},
				},
			},
			{
				Code: `const obj = { a: 1 }; enum InvalidSpread { A = 'TestStr', B = { ...a }, }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 60},
				},
			},
			{
				Code:    `enum Foo { A = 1 << 0, B = 1 >> 0, C = 1 >>> 0, D = 1 | 0, E = 1 & 0, F = 1 ^ 0, G = ~1, }`,
				Options: map[string]interface{}{"allowBitwiseExpressions": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 12},
					{MessageId: "notLiteral", Line: 1, Column: 24},
					{MessageId: "notLiteral", Line: 1, Column: 36},
					{MessageId: "notLiteral", Line: 1, Column: 50},
					{MessageId: "notLiteral", Line: 1, Column: 61},
					{MessageId: "notLiteral", Line: 1, Column: 72},
					{MessageId: "notLiteral", Line: 1, Column: 83},
				},
			},
			{
				Code:    `const x = 1; enum Foo { A = x << 0, B = x >> 0, C = x >>> 0, D = x | 0, E = x & 0, F = x ^ 0, G = ~x, }`,
				Options: map[string]interface{}{"allowBitwiseExpressions": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteralOrBitwiseExpression", Line: 1, Column: 25},
					{MessageId: "notLiteralOrBitwiseExpression", Line: 1, Column: 38},
					{MessageId: "notLiteralOrBitwiseExpression", Line: 1, Column: 50},
					{MessageId: "notLiteralOrBitwiseExpression", Line: 1, Column: 63},
					{MessageId: "notLiteralOrBitwiseExpression", Line: 1, Column: 74},
					{MessageId: "notLiteralOrBitwiseExpression", Line: 1, Column: 85},
					{MessageId: "notLiteralOrBitwiseExpression", Line: 1, Column: 97},
				},
			},
			{
				Code:    `const x = 1; enum Foo { A = 1 << 0, B = x >> Foo.A, C = x >> A, }`,
				Options: map[string]interface{}{"allowBitwiseExpressions": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteralOrBitwiseExpression", Line: 1, Column: 38},
					{MessageId: "notLiteralOrBitwiseExpression", Line: 1, Column: 54},
				},
			},
			{
				Code: `enum Foo { A, B = +A, }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLiteral", Line: 1, Column: 15},
				},
			},
		},
	)
}
