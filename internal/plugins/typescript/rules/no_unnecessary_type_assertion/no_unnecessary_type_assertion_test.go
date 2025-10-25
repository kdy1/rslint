package no_unnecessary_type_assertion

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoUnnecessaryTypeAssertionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryTypeAssertionRule,
		[]rule_tester.ValidTestCase{
			// Valid code without assertions
			{Code: "const x = 1;"},
			{Code: "const y: number = 2;"},
			{Code: "function foo(x: number) { return x; }"},

			// Necessary type narrowing
			{Code: "const x: any = 1;\nconst y = x as number;"},
			{Code: "const x: unknown = 'hello';\nconst y = x as string;"},

			// Non-null assertions on nullable types (handled by this rule but valid)
			{Code: "const x: number | null = 1;\nconst y = x!;"},
		},
		[]rule_tester.InvalidTestCase{
			// Unnecessary number assertion
			{
				Code: "const x = 3 as 3;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssertion", Line: 1, Column: 11},
				},
				Output: []string{"const x = 3;"},
			},

			// Unnecessary assertion on already typed variable
			{
				Code: "const x: number = 1;\nconst y = x as number;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryAssertion", Line: 2, Column: 11},
				},
				Output: []string{"const x: number = 1;\nconst y = x;"},
			},
		},
	)
}
