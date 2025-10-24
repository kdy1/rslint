package no_loss_of_precision

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoLossOfPrecisionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLossOfPrecisionRule,
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
const x = 9007199254740993;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "default",
						Line:      2,
						Column:    11,
					},
				},
			},
		},
	)
}
