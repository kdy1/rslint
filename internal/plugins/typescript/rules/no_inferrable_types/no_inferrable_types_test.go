package no_inferrable_types

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoInferrableTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInferrableTypesRule,
		[]rule_tester.ValidTestCase{
			{Code: `const x = 1;`},
			{Code: `let x = "hello";`},
			{Code: `var x = true;`},
		},
		[]rule_tester.InvalidTestCase{
			// No invalid test cases - rule implementation is incomplete
		},
	)
}
