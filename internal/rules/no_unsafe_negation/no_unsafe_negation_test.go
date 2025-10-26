package no_unsafe_negation

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoUnsafeNegationRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnsafeNegationRule,
		[]rule_tester.ValidTestCase{
			// TODO: Add valid test cases
			{Code: `
// Add valid code example here
const x = 1;
`},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases
			{
				Code: `
// Add invalid code example here
var x = 1;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      2, // TODO: Update line number
						Column:    1, // TODO: Update column number
					},
				},
				Output: []string{`
// Add expected output after autofix
const x = 1;
`},
			},
		},
	)
}

func TestNoUnsafeNegationRuleWithOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnsafeNegationRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `
// Add code that is valid with specific options
`,
				Options: map[string]interface{}{
					// TODO: Add option values
					// "optionName": true,
				},
			},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
