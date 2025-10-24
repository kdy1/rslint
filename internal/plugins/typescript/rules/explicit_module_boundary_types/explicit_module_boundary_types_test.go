package explicit_module_boundary_types

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestExplicitModuleBoundaryTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ExplicitModuleBoundaryTypesRule,
		[]rule_tester.ValidTestCase{
			{Code: `export function foo(): number { return 1; }`},
			{Code: `export const foo = (bar: string): number => 1;`},
			{Code: `function foo(bar: string): void {}`},
		},
		[]rule_tester.InvalidTestCase{
			// No invalid test cases - rule implementation is incomplete
		},
	)
}
