package no_new

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewRule,
		[]rule_tester.ValidTestCase{
			// Assignment of newly instantiated object to variable
			{Code: `var a = new Date()`},

			// Using new expression in conditional comparison
			{Code: `var a; if (a === new Date()) { a = false; }`},
		},
		[]rule_tester.InvalidTestCase{
			// Standalone constructor call without assignment
			{
				Code: `new Date()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noNewStatement",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}
