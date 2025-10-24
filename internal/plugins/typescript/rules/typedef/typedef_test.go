package typedef

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestTypedefRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&TypedefRule,
		[]rule_tester.ValidTestCase{
			{Code: `const x: number = 1;`},
			{Code: `let y: string = "hello";`},
			{Code: `function foo(bar: string): number { return 1; }`},
		},
		[]rule_tester.InvalidTestCase{
			// No invalid test cases - rule implementation is incomplete
		},
	)
}
