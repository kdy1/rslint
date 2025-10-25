package prefer_const

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestPreferConstRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferConstRule,
		[]rule_tester.ValidTestCase{
			// Variables that are reassigned
			{Code: "let x = 0; x = 1;"},
			// Loop variables that are modified
			{Code: "for (let i in [1,2,3]) { i = 0; }"},
			// Variables assigned after declaration
			{Code: "let x; x = 0;"},
			// const declarations
			{Code: "const x = 1;"},
			// var declarations (not checked by this rule)
			{Code: "var x = 1;"},
		},
		[]rule_tester.InvalidTestCase{
			// Simple case
			{
				Code: "let x = 1; foo(x);",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useConst",
					},
				},
				Output: []string{"const x = 1; foo(x);"},
			},
			// Multiple declarations (simplified - only checks if initialized)
			{
				Code: "let x = 1, y = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useConst",
					},
				},
				Output: []string{"const x = 1, y = 2;"},
			},
			// Destructuring
			{
				Code: "let {a, b} = obj;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useConst",
					},
				},
				Output: []string{"const {a, b} = obj;"},
			},
			// Array destructuring
			{
				Code: "let [a, b] = arr;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useConst",
					},
				},
				Output: []string{"const [a, b] = arr;"},
			},
		},
	)
}
