package class_methods_use_this

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestClassMethodsUseThisRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ClassMethodsUseThisRule,
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
class A {
  method() {
    return 42;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingThis",
						Line:      3,
						Column:    3,
					},
				},
			},
		},
	)
}
