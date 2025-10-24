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
		[]rule_tester.InvalidTestCase{
			{
				Code: `
const x: string = "hello";
if (x) {
  console.log('Invalid');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "conditionErrorString",
						Line:      3,
						Column:    5,
					},
				},
			},
			{
				Code: `
const x: number = 42;
if (x) {
  console.log('Invalid');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "conditionErrorNumber",
						Line:      3,
						Column:    5,
					},
				},
			},
			{
				Code: `
const x: object | null = {};
if (x) {
  console.log('Invalid');
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "conditionErrorNullableObject",
						Line:      3,
						Column:    5,
					},
				},
			},
		},
	)
}
