package array_callback_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestArrayCallbackReturnRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ArrayCallbackReturnRule,
		[]rule_tester.ValidTestCase{
			// TODO: Add valid test cases when rule logic is implemented
			{Code: `const x = 1;`},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases when rule logic is implemented
		},
	)
}
