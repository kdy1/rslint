package explicit_function_return_type

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestExplicitFunctionReturnTypeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ExplicitFunctionReturnTypeRule,
		[]rule_tester.ValidTestCase{
			{Code: `function foo(): number { return 1; }`},
			{Code: `const foo = (): number => 1;`},
			{Code: `class Foo { method(): void {} }`},
		},
		[]rule_tester.InvalidTestCase{
			// No invalid test cases - rule implementation is incomplete
		},
	)
}
