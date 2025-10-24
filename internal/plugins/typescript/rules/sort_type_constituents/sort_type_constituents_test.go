package sort_type_constituents

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortTypeConstituentsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SortTypeConstituentsRule,
		[]rule_tester.ValidTestCase{
			{Code: `type T = A | B | C;`},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases
		},
	)
}
