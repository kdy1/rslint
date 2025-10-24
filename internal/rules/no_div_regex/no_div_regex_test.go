package no_div_regex

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoDivRegexRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDivRegexRule,
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
			},
		},
	)
}

func TestNoDivRegexRuleWithOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDivRegexRule,
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
