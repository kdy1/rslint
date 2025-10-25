package prefer_destructuring

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestPreferDestructuringRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		[]rule_tester.ValidTestCase{
			{Code: `var [foo] = array;`},
			{Code: `var { foo } = object;`},
			{Code: `var foo;`},
			{Code: `var foo = object?.foo;`}, // Optional chaining is exempt
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `var foo = array[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
			},
			{
				Code: `var foo = object.foo;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
				Output: []string{`var {foo} = object;`},
			},
			{
				Code: `var foo = object.bar.foo;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
				Output: []string{`var {foo} = object.bar;`},
			},
			{
				Code: `foo = object.foo;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
			},
		},
	)
}

func TestPreferDestructuringRuleWithOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    `var foo = array[0];`,
				Options: map[string]interface{}{"array": false},
			},
			{
				Code:    `var foo = object.foo;`,
				Options: map[string]interface{}{"object": false},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `var foo = array[0];`,
				Options: map[string]interface{}{"array": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
			},
		},
	)
}
