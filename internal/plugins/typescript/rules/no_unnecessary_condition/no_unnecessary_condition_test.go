package no_unnecessary_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryConditionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryConditionRule,
		[]rule_tester.ValidTestCase{
			{Code: `
const x: string | undefined = getX();
if (x) {
  console.log(x);
}
`},
			{Code: `
const value: number | null = getValue();
const result = value ?? 0;
`},
			{Code: `
while (true) {
  break;
}
`, Options: map[string]interface{}{"allowConstantLoopConditions": true}},
		},
		// TODO: Add invalid test cases once rule implementation is complete
		[]rule_tester.InvalidTestCase{},
	)
}
