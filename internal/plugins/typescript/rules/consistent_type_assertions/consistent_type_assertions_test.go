package consistent_type_assertions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTypeAssertions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentTypeAssertionsRule,
		[]rule_tester.ValidTestCase{
			// Default: prefer 'as' style
			{Code: `const x = value as string;`},
			{Code: `const y = obj as MyType;`},

			// Const assertions are always allowed
			{Code: `const x = "hello" as const;`},
			{Code: `const obj = { x: 1 } as const;`},

			// Angle-bracket preference
			{
				Code: `const x = <string>value;`,
				Options: map[string]interface{}{
					"assertionStyle": "angle-bracket",
				},
			},

			// Object literals with allow
			{
				Code: `const x = { a: 1 } as MyType;`,
				Options: map[string]interface{}{
					"objectLiteralTypeAssertions": "allow",
				},
			},

			// Assertions to any/unknown bypass restrictions
			{
				Code: `const x = { a: 1 } as any;`,
				Options: map[string]interface{}{
					"assertionStyle":               "never",
					"objectLiteralTypeAssertions": "never",
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Prefer 'as' (default): angle-bracket used
			{
				Code: `const x = <string>value;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "as"},
				},
			},

			// Prefer angle-bracket: 'as' used
			{
				Code: `const x = value as string;`,
				Options: map[string]interface{}{
					"assertionStyle": "angle-bracket",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "angle-bracket"},
				},
			},

			// Never: disallow all assertions
			{
				Code: `const x = value as string;`,
				Options: map[string]interface{}{
					"assertionStyle": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "never"},
				},
			},

			// Object literal never
			{
				Code: `const x = { a: 1 } as MyType;`,
				Options: map[string]interface{}{
					"objectLiteralTypeAssertions": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedObjectTypeAssertion"},
				},
			},

			// Array literal never
			{
				Code: `const x = [1, 2, 3] as number[];`,
				Options: map[string]interface{}{
					"arrayLiteralTypeAssertions": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedArrayTypeAssertion"},
				},
			},
		},
	)
}
