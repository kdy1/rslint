package no_non_null_assertion

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoNonNullAssertionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNonNullAssertionRule,
		[]rule_tester.ValidTestCase{
			// Simple variable references without non-null assertion
			{Code: "const x = 1;"},
			{Code: "const y = x;"},
			{Code: "x.y;"},
			{Code: "x.y.z;"},

			// Optional chaining (valid alternative)
			{Code: "x?.y;"},
			{Code: "x?.y?.z;"},
			{Code: "x?.[y];"},

			// Logical NOT (single negation is valid)
			{Code: "if (!x) {}"},
			{Code: "const result = !value;"},
		},
		[]rule_tester.InvalidTestCase{
			// Basic non-null assertions
			{
				Code: "x!;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNonNullAssertion", Line: 1, Column: 1},
				},
			},
			{
				Code: "x.y!;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNonNullAssertion", Line: 1, Column: 1},
				},
			},

			// Non-null assertion with property access
			{
				Code: "x!.y;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNonNullAssertion", Line: 1, Column: 1},
				},
			},

			// Non-null assertion with element access
			{
				Code: "x![y];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNonNullAssertion", Line: 1, Column: 1},
				},
			},

			// Non-null assertion with function call
			{
				Code: "x.y.z!();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNonNullAssertion", Line: 1, Column: 1},
				},
			},

			// Multiple non-null assertions
			{
				Code: "x!!!;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNonNullAssertion", Line: 1, Column: 1},
					{MessageId: "noNonNullAssertion", Line: 1, Column: 1},
					{MessageId: "noNonNullAssertion", Line: 1, Column: 1},
				},
			},
		},
	)
}
