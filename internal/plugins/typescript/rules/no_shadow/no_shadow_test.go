package no_shadow

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoShadowRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoShadowRule,
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
const x = 1;
function foo() {
  const x = 2;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noShadow",
						Line:      4,
						Column:    9,
					},
				},
			},
		},
	)
}
