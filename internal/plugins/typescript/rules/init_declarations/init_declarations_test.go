package init_declarations

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestInitDeclarationsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&InitDeclarationsRule,
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
var x;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "default",
						Line:      2,
						Column:    5,
					},
				},
			},
		},
	)
}
