package strict_boolean_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStrictBooleanExpressionsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&StrictBooleanExpressionsRule,
		[]rule_tester.ValidTestCase{
			{Code: `
const x: boolean = true;
if (x) {
  console.log('Valid');
}
`},
			{Code: `
const x: string = "hello";
if (x) {
  console.log('Allowed');
}
`, Options: map[string]interface{}{"allowString": true}},
			{Code: `
const x: boolean = Math.random() > 0.5;
while (x) {
  break;
}
`},
		},
		// TODO: Add invalid test cases once rule implementation is complete
		[]rule_tester.InvalidTestCase{},
	)
}
