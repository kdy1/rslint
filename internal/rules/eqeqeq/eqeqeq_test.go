package eqeqeq

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestEqeqeqRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &EqeqeqRule, []rule_tester.ValidTestCase{
		// Valid cases - default "always" mode
		{Code: "a === b"},
		{Code: "a !== b"},
		{Code: "typeof a === 'number'"},
		{Code: "typeof a !== 'number'"},
		{Code: "'hello' === 'world'"},
		{Code: "true === true"},
		{Code: "null === null"},

		// Valid cases - smart mode
		{Code: "typeof a == 'number'", Options: []interface{}{"smart"}},
		{Code: "'hello' == 'world'", Options: []interface{}{"smart"}},
		{Code: "0 == 0", Options: []interface{}{"smart"}},
		{Code: "a == null", Options: []interface{}{"smart"}},
		{Code: "null == a", Options: []interface{}{"smart"}},

		// Valid cases - null: ignore
		{Code: "a == null", Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
		{Code: "null == a", Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
		{Code: "a != null", Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},

		// Valid cases - null: never
		{Code: "a == null", Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
		{Code: "null == a", Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases - default "always" mode
		{
			Code: "a == b",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
			Output: []string{"a === b"},
		},
		{
			Code: "a != b",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
			Output: []string{"a !== b"},
		},
		{
			Code: "typeof a == 'number'",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
			Output: []string{"typeof a === 'number'"},
		},
		{
			Code: "a == null",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
			Output: []string{"a === null"},
		},

		// Invalid cases - smart mode
		{
			Code:    "a == b",
			Options: []interface{}{"smart"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
			Output: []string{"a === b"},
		},
		{
			Code:    "foo.bar == true",
			Options: []interface{}{"smart"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
			Output: []string{"foo.bar === true"},
		},

		// Invalid cases - null: never (should use == for null)
		{
			Code:    "a === null",
			Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNull"},
			},
			Output: []string{"a == null"},
		},
		{
			Code:    "null !== a",
			Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedNull"},
			},
			Output: []string{"null != a"},
		},
	})
}
