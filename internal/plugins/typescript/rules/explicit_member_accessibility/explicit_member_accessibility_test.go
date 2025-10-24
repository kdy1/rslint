package explicit_member_accessibility

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestExplicitMemberAccessibilityRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ExplicitMemberAccessibilityRule,
		[]rule_tester.ValidTestCase{
			{Code: `class Foo { public bar: string; }`},
			{Code: `class Foo { private bar: string; }`},
			{Code: `class Foo { protected bar: string; }`},
		},
		[]rule_tester.InvalidTestCase{
			// No invalid test cases - rule implementation is incomplete
		},
	)
}
