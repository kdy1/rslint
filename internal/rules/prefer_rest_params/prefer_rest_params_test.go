package prefer_rest_params

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestPreferRestParamsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferRestParamsRule,
		[]rule_tester.ValidTestCase{
			{Code: `arguments;`},
			{Code: `function foo(arguments) { arguments; }`},
			{Code: `function foo() { var arguments; arguments; }`},
			{Code: `var foo = () => arguments;`},
			{Code: `function foo(...args) { args; }`},
			{Code: `function foo() { arguments.length; }`},
			{Code: `function foo() { arguments.callee; }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `function foo() { arguments; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams"},
				},
			},
			{
				Code: `function foo() { arguments[0]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams"},
				},
			},
			{
				Code: `function foo() { arguments[1]; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferRestParams"},
				},
			},
		},
	)
}
