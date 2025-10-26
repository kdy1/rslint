package no_caller

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCallerRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoCallerRule,
		[]rule_tester.ValidTestCase{
			// Valid: accessing arguments.length
			{Code: `var x = arguments.length`},

			// Valid: accessing arguments object itself
			{Code: `var x = arguments`},

			// Valid: accessing arguments by index
			{Code: `var x = arguments[0]`},

			// Valid: accessing arguments with bracket notation (caller is a variable, not the property)
			{Code: `var x = arguments[caller]`},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid: accessing arguments.callee
			{
				Code: `var x = arguments.callee`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    9,
					},
				},
			},

			// Invalid: accessing arguments.caller
			{
				Code: `var x = arguments.caller`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    9,
					},
				},
			},
		},
	)
}
