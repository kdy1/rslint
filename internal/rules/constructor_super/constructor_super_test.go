package constructor_super

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConstructorSuperRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConstructorSuperRule,
		[]rule_tester.ValidTestCase{
			// TODO: Add valid test cases when rule logic is implemented
			{Code: `const x = 1;`},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases when rule logic is implemented
		},
	)
}
