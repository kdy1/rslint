package restrict_template_expressions

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestRestrictTemplateExpressionsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RestrictTemplateExpressionsRule,
		[]rule_tester.ValidTestCase{
			// TODO: Add valid test cases
			{Code: `
// Add valid code example here
const x = 1;
`},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases when the rule implementation is complete
		},
	)
}
