package eqeqeq

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestEqeqeqRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&EqeqeqRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Default "always" mode
			{Code: `a === b`},
			{Code: `a !== b`},

			// With explicit "always" option
			{
				Code:    `a === b`,
				Options: []interface{}{"always"},
			},
			{
				Code:    `a !== b`,
				Options: []interface{}{"always"},
			},

			// Smart mode - allows == for typeof
			{
				Code:    `typeof a == 'number'`,
				Options: []interface{}{"smart"},
			},
			{
				Code:    `'string' != typeof a`,
				Options: []interface{}{"smart"},
			},

			// Smart mode - allows == for matching literals
			{
				Code:    `'hello' != 'world'`,
				Options: []interface{}{"smart"},
			},
			{
				Code:    `2 == 3`,
				Options: []interface{}{"smart"},
			},
			{
				Code:    `true == true`,
				Options: []interface{}{"smart"},
			},

			// Smart mode - allows null comparisons
			{
				Code:    `null == a`,
				Options: []interface{}{"smart"},
			},
			{
				Code:    `a == null`,
				Options: []interface{}{"smart"},
			},

			// Allow-null option
			{
				Code:    `null == a`,
				Options: []interface{}{"allow-null"},
			},
			{
				Code:    `a == null`,
				Options: []interface{}{"allow-null"},
			},
			{
				Code:    `a === b`,
				Options: []interface{}{"allow-null"},
			},

			// Always with null: "ignore"
			{
				Code:    `a == null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}},
			},
			{
				Code:    `a != null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}},
			},

			// Always with null: "always"
			{
				Code:    `a === null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "always"}},
			},
			{
				Code:    `a !== null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "always"}},
			},

			// Always with null: "never"
			{
				Code:    `null == null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
			},
			{
				Code:    `a == null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
			},

			// BigInt literals
			{
				Code:    `foo === 1n`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
			},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Default "always" mode violations
			{
				Code: `a == b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected"},
				},
			},
			{
				Code: `a != b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedNot"},
				},
			},

			// Smart mode - type mismatches
			{
				Code:    `true == 1`,
				Options: []interface{}{"smart"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected"},
				},
			},
			{
				Code:    `0 != '1'`,
				Options: []interface{}{"smart"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedNot"},
				},
			},
			{
				Code:    `a == b`,
				Options: []interface{}{"smart"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected"},
				},
			},
			{
				Code:    `foo == true`,
				Options: []interface{}{"smart"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected"},
				},
			},

			// null: "always" violations
			{
				Code:    `true == null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected"},
				},
			},
			{
				Code:    `null == null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected"},
				},
			},

			// null: "never" violations
			{
				Code:    `true === null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code:    `null !== true`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedNot"},
				},
			},

			// Mixed cases
			{
				Code: `x == 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected"},
				},
			},
			{
				Code: `x != 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedNot"},
				},
			},

			// String comparisons with variables
			{
				Code:    `foo == "test"`,
				Options: []interface{}{"smart"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected"},
				},
			},

			// Number comparisons with variables
			{
				Code:    `foo == 42`,
				Options: []interface{}{"smart"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected"},
				},
			},

			// Boolean comparisons with variables
			{
				Code:    `foo == false`,
				Options: []interface{}{"smart"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expected"},
				},
			},
		},
	)
}
