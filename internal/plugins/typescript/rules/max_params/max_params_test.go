package max_params

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestMaxParamsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&MaxParamsRule,
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
function foo(a, b, c, d) {
  return a + b + c + d;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "default",
						Line:      2,
						Column:    1,
					},
				},
			},
		},
	)
}
