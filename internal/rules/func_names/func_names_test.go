package func_names

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestFuncNamesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&FuncNamesRule,
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
						MessageId: "unnamed",
						Line:      2, // TODO: Update line number
						Column:    1, // TODO: Update column number
					},
				},
			},
		},
	)
}

func TestFuncNamesRuleWithOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&FuncNamesRule,
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
