package valid_typeof

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestValidTypeofRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ValidTypeofRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Basic valid string comparisons
			{Code: `typeof foo === "string"`},
			{Code: `typeof foo === "object"`},
			{Code: `typeof foo === "function"`},
			{Code: `typeof foo === "undefined"`},
			{Code: `typeof foo === "boolean"`},
			{Code: `typeof foo === "number"`},
			{Code: `typeof foo === "bigint"`},
			{Code: `typeof foo === "symbol"`},

			// Reversed comparisons
			{Code: `"string" === typeof foo`},
			{Code: `"object" === typeof foo`},
			{Code: `"function" === typeof foo`},
			{Code: `"undefined" === typeof foo`},
			{Code: `"boolean" === typeof foo`},
			{Code: `"number" === typeof foo`},

			// typeof-to-typeof comparisons
			{Code: `typeof foo === typeof bar`},
			{Code: `typeof foo === baz`},
			{Code: `typeof foo !== someType`},
			{Code: `typeof bar != someType`},
			{Code: `someType === typeof bar`},
			{Code: `someType == typeof bar`},

			// Different operators
			{Code: `typeof foo == "string"`},
			{Code: `typeof(foo) === "string"`},
			{Code: `typeof(foo) !== "string"`},
			{Code: `typeof(foo) == "string"`},
			{Code: `typeof(foo) != "string"`},

			// Non-comparison uses
			{Code: `var oddUse = typeof foo + "thing"`},
			{Code: `function f(undefined) { typeof x === undefined }`},

			// Template literals
			{Code: "typeof foo === `string`"},
			{Code: "`object` === typeof foo"},

			// With requireStringLiterals option
			{
				Code:    `typeof foo === "number"`,
				Options: []interface{}{map[string]interface{}{"requireStringLiterals": true}},
			},
			{
				Code:    `typeof foo === "string"`,
				Options: []interface{}{map[string]interface{}{"requireStringLiterals": true}},
			},
			{
				Code:    `var baz = typeof foo + "thing"`,
				Options: []interface{}{map[string]interface{}{"requireStringLiterals": true}},
			},
			{
				Code:    `typeof foo === typeof bar`,
				Options: []interface{}{map[string]interface{}{"requireStringLiterals": true}},
			},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Invalid string values (typos)
			{
				Code: `typeof foo === "strnig"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue"},
				},
			},
			{
				Code: `"strnig" === typeof foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue"},
				},
			},
			{
				Code: `if (typeof bar === "umdefined") {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue"},
				},
			},
			{
				Code: `typeof foo !== "strnig"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue"},
				},
			},
			{
				Code: `typeof foo != "strnig"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue"},
				},
			},
			{
				Code: `typeof foo == "strnig"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue"},
				},
			},
			{
				Code: "if (typeof bar == `umdefined`) {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue"},
				},
			},

			// With requireStringLiterals option
			{
				Code:    `typeof foo === undefined`,
				Options: []interface{}{map[string]interface{}{"requireStringLiterals": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notString"},
				},
			},
			{
				Code:    `undefined === typeof foo`,
				Options: []interface{}{map[string]interface{}{"requireStringLiterals": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notString"},
				},
			},
			{
				Code:    `undefined == typeof foo`,
				Options: []interface{}{map[string]interface{}{"requireStringLiterals": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notString"},
				},
			},
			{
				Code:    `typeof foo === Object`,
				Options: []interface{}{map[string]interface{}{"requireStringLiterals": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notString"},
				},
			},
			{
				Code:    "typeof foo === `undefined${foo}`",
				Options: []interface{}{map[string]interface{}{"requireStringLiterals": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notString"},
				},
			},
		},
	)
}
