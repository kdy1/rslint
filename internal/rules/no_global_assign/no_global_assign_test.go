package no_global_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoGlobalAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoGlobalAssignRule,
		[]rule_tester.ValidTestCase{
			// Assigning to local variables is allowed
			{Code: `string = "hello world";`},
			{Code: `var string;`},
			{Code: `top = 0;`},
			{Code: `onload = 0;`},
		},
		[]rule_tester.InvalidTestCase{
			// Assigning to built-in globals is not allowed
			{
				Code: `String = "hello world";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified"},
				},
			},
			{
				Code: `String++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified"},
				},
			},
			{
				Code: `Array = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified"},
				},
			},
			{
				Code: `Object = 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified"},
				},
			},
			{
				Code: `undefined = true;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified"},
				},
			},
			{
				Code: `NaN = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified"},
				},
			},
			{
				Code: `Infinity = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified"},
				},
			},
			{
				Code: `Math = {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified"},
				},
			},
		},
	)
}

func TestNoGlobalAssignRuleWithOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoGlobalAssignRule,
		[]rule_tester.ValidTestCase{
			// With exceptions option, specified globals can be assigned
			{
				Code: `Object = 0;`,
				Options: map[string]interface{}{
					"exceptions": []interface{}{"Object"},
				},
			},
			{
				Code: `String = "test";`,
				Options: map[string]interface{}{
					"exceptions": []interface{}{"String"},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Without exceptions, these should still fail
			{
				Code: `Array = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "globalShouldNotBeModified"},
				},
			},
		},
	)
}
