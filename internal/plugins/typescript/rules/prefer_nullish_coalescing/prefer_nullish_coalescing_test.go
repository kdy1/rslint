package prefer_nullish_coalescing

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestPreferNullishCoalescingRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferNullishCoalescingRule,
		[]rule_tester.ValidTestCase{
			// Stub tests - rule is not fully implemented
			{Code: `const x = a ?? b;`},
			{Code: `const x = a !== null && a !== undefined ? a : b;`, Options: map[string]interface{}{"ignoreTernaryTests": true}},
		},
		[]rule_tester.InvalidTestCase{
			// No invalid cases for stub implementation
		},
	)
}
