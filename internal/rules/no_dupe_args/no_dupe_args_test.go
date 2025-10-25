package no_dupe_args

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeArgsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeArgsRule,
		[]rule_tester.ValidTestCase{
			// Valid cases - no duplicate arguments
			{Code: "function a(a, b, c){}"},
			{Code: "var a = function(a, b, c){}"},
			{Code: "function a({a, b}, {c, d}){}"},
			{Code: "function a([ , a]) {}"},
			{Code: "function foo([[a, b], [c, d]]) {}"},
			{Code: "function a(a: number, b: string, c: boolean) {}"},
			{Code: "const fn = (x, y, z) => {}"},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid cases - duplicate arguments
			{
				Code: "function a(a, b, b) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    18,
					},
				},
			},
			{
				Code: "function a(a, a, a) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    15,
					},
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    18,
					},
				},
			},
			{
				Code: "function a(a, b, a) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    18,
					},
				},
			},
			{
				Code: "function a(a, b, a, b) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    18,
					},
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    21,
					},
				},
			},
			{
				Code: "var a = function(a, b, b) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    24,
					},
				},
			},
			{
				Code: "var a = function(a, a, a) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    21,
					},
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    24,
					},
				},
			},
			{
				Code: "var a = function(a, b, a) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    24,
					},
				},
			},
			{
				Code: "var a = function(a, b, a, b) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    24,
					},
					{
						MessageId: "duplicateParam",
						Line:      1,
						Column:    27,
					},
				},
			},
		},
	)
}
