package no_misleading_character_class

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoMisleadingCharacterClassRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoMisleadingCharacterClassRule,
		[]rule_tester.ValidTestCase{
			// With unicode flag
			{Code: "var r = /[👍]/u;"},
			{Code: "var r = /[\\uD83D\\uDC4D]/u;"},

			// Outside character class
			{Code: "var r = /❇️/;"},

			// Empty RegExp
			{Code: "new RegExp();"},

			// Solo characters (not combined)
			{Code: "var r = /[\\uD83D]/;"},
			{Code: "var r = /[\\u0301]/;"},

			// Single letters
			{Code: "var r = /[JP]/;"},

			// Basic patterns
			{Code: "var r = /[abc]/;"},
			{Code: "var r = /[a-z]/;"},
		},
		[]rule_tester.InvalidTestCase{
			// Surrogate pair without u flag
			{
				Code: "var r = /[👍]/;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "surrogatePairWithoutUFlag"},
				},
			},

			// More complex invalid cases can be added here
			// Note: Testing all Unicode categories would require more sophisticated test data
		},
	)
}
