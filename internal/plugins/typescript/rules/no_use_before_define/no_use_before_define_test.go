package no_use_before_define

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoUseBeforeDefineRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUseBeforeDefineRule,
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
						MessageId: "noUseBeforeDefine",
						Line:      2, // TODO: Update line number
						Column:    1, // TODO: Update column number
					},
				},
			},
		},
	)
}
