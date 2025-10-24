package no_unnecessary_qualifier

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryQualifierRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryQualifierRule,
		[]rule_tester.ValidTestCase{
			{Code: `
namespace N {
  export const x = 1;
}
const y = N.x;
`},
			{Code: `
enum E {
  A,
  B
}
const value = E.A;
`},
		},
		// TODO: Add invalid test cases once rule implementation is complete
		[]rule_tester.InvalidTestCase{},
	)
}
