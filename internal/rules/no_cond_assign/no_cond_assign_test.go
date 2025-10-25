package no_cond_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCondAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoCondAssignRule,
		// Valid test cases
		[]rule_tester.ValidTestCase{
			// Default mode: except-parens
			{Code: "const x = 0;"},
			{Code: "const x = (y = 0);"},
			{Code: "if (x === 0) { }"},
			{Code: "if ((x = y)) { }"},
			{Code: "while ((a = b));"},
			{Code: "do { } while ((a = b));"},
			{Code: "for (;(a = b););"},
			{Code: "for (;;) { }"},
			{Code: "if (someNode || (someNode = parentNode)) { }"},
			{Code: "while (someNode || (someNode = parentNode)) { }"},
			{Code: "do { } while (someNode || (someNode = parentNode));"},
			{Code: "for (;someNode || (someNode = parentNode););"},
			{Code: "if ((function(node) { return node = parentNode; })(someNode)) { }"},
			{Code: "if ((function(node) { return node = parentNode; })) { }"},
			{Code: "if ((node => node = parentNode)) { }"},
			{Code: "const x = (a ? (b = c) : d);"},
			{Code: "while ((x = y) !== 0) { }"},

			// Test always mode
			{Code: "const x = 0;", Options: "always"},
			{Code: "if (x === 0) { }", Options: "always"},
			{Code: "while (x) { }", Options: "always"},
			{Code: "do { } while (x);", Options: "always"},
		},
		// Invalid test cases
		[]rule_tester.InvalidTestCase{
			// Default mode: except-parens
			{
				Code: "if (x = 0) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
				},
			},
			{
				Code: "while (x = 0) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
				},
			},
			{
				Code: "do { } while (x = 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
				},
			},
			{
				Code: "for (;x = 0;) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
				},
			},
			{
				Code: "if (x = 0) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
				},
			},
			{
				Code: "while (x = 1) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
				},
			},
			{
				Code: "do { } while (x = 1);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
				},
			},
			{
				Code: "const x = (a) ? (b = c) : d;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
				},
			},
			{
				Code: "for (;x = 1;) { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
				},
			},
			{
				Code: "const x = (a = b) ? c : d;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing"},
				},
			},

			// Always mode
			{
				Code:    "if ((x = 0)) { }",
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code:    "while ((x = 0)) { }",
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code:    "do { } while ((x = 0));",
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code:    "for (;(x = 0);) { }",
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code:    "if (x = 0) { }",
				Options: "always",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
		},
	)
}
