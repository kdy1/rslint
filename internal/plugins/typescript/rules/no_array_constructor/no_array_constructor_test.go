package no_array_constructor

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoArrayConstructorRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoArrayConstructorRule,
		[]rule_tester.ValidTestCase{
			// TODO: Add valid test cases
			{Code: `
// Add valid code example here
const x = 1;
`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
const arr = new Array(1, 2, 3);
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useLiteral",
						Line:      2,
						Column:    13,
					},
				},
			},
		},
	)
}
