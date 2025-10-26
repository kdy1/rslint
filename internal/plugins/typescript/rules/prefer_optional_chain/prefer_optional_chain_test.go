package prefer_optional_chain

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestPreferOptionalChainRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferOptionalChainRule,
		[]rule_tester.ValidTestCase{
			// Stub tests - rule is not fully implemented
			{Code: `foo?.bar?.baz`},
			{Code: `foo?.[bar]`},
			{Code: `foo?.bar?.()`},
		},
		[]rule_tester.InvalidTestCase{
			// No invalid cases for stub implementation
		},
	)
}
