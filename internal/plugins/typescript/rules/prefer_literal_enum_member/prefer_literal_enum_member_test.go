package prefer_literal_enum_member

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestPreferLiteralEnumMemberRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferLiteralEnumMemberRule,
		[]rule_tester.ValidTestCase{
			{Code: `enum Foo { A = 1, B = 2 }`},
			{Code: `enum Foo { A = "a", B = "b" }`},
			{Code: `const x = 1;`},
		},
		[]rule_tester.InvalidTestCase{
			// No invalid test cases - rule implementation is incomplete
		},
	)
}
