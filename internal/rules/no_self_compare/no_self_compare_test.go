package no_self_compare

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSelfCompareRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoSelfCompareRule,
		[]rule_tester.ValidTestCase{
			// Valid comparisons
			{Code: "if (x === y) { }"},
			{Code: "if (1 === 2) { }"},
			{Code: "y = x * x;"},
			{Code: "foo.bar.baz === foo.bar.qux;"},
		},
		[]rule_tester.InvalidTestCase{
			// Equality operators
			{
				Code: "if (x === x) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},
			{
				Code: "if (x !== x) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},
			{
				Code: "if (x == x) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},
			{
				Code: "if (x != x) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},

			// Relational operators
			{
				Code: "if (x > x) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},
			{
				Code: "if (x < x) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},
			{
				Code: "if (x >= x) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},
			{
				Code: "if (x <= x) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},

			// String literals
			{
				Code: "if ('x' > 'x') { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},

			// Property access
			{
				Code: "if (foo.bar.baz === foo.bar.baz) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},
			{
				Code: "foo.bar().baz.qux >= foo.bar ().baz .qux;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},

			// In control structures
			{
				Code: "do {} while (x === x);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparingToSelf"},
				},
			},
		})
}
