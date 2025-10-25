package no_dupe_else_if

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeElseIfRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeElseIfRule,
		[]rule_tester.ValidTestCase{
			// Valid cases - no duplicate conditions
			{Code: "if (a) {} else if (b) {}"},
			{Code: "if (a) {} else if (b) {} else if (c) {}"},
			{Code: "if (true) {} else if (false) {} else {}"},
			{Code: "if (1) {} else if (2) {}"},
			{Code: "if (f) {} else if (f()) {}"},
			{Code: "if (f(a)) {} else if (g(a)) {}"},
			{Code: "if (f(a)) {} else if (f(b)) {}"},
			{Code: "if (a === 1) {} else if (a === 2) {}"},
			{Code: "if (a === 1) {} else if (b === 1) {}"},
			{Code: "if (a) {}"},
			{Code: "if (a) {} else {}"},
			{Code: "if (a || b) {} else if (c || d) {}"},
			{Code: "if (a || b) {} else if (a || c) {}"},
			{Code: "if (a) {} else if (a || b) {}"},
			{Code: "if (a && b) {} else if (a) {} else if (b) {}"},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid cases - duplicate conditions
			{
				Code: "if (a) {} else if (a) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateCondition",
						Line:      1,
						Column:    20,
					},
				},
			},
			{
				Code: "if (a) {} else if (a) {} else {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateCondition",
						Line:      1,
						Column:    20,
					},
				},
			},
			{
				Code: "if (a) {} else if (a) {} else if (a) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateCondition",
						Line:      1,
						Column:    20,
					},
					{
						MessageId: "duplicateCondition",
						Line:      1,
						Column:    35,
					},
				},
			},
			{
				Code: "if (a === 1) {} else if (a === 1) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateCondition",
						Line:      1,
						Column:    26,
					},
				},
			},
			{
				Code: "if (a && b) {} else if (a && b) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateCondition",
						Line:      1,
						Column:    25,
					},
				},
			},
		},
	)
}
