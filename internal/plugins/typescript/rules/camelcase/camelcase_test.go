package camelcase

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCamelcaseRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&CamelcaseRule,
		[]rule_tester.ValidTestCase{
			{Code: `const myVariable = 1;`},
			{Code: `function myFunction() {}`},
			{Code: `class MyClass {}`},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add comprehensive invalid test cases
		},
	)
}
