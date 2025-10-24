package no_redeclare

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoRedeclareRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRedeclareRule,
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
var x = 1;
var x = 2;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redeclared",
						Line:      3,
						Column:    5,
					},
				},
			},
		},
	)
}
