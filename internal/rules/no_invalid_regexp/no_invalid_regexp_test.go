package no_invalid_regexp

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidRegexpRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInvalidRegexpRule, []rule_tester.ValidTestCase{
		// Valid: empty and simple patterns
		{Code: "RegExp('')"},
		{Code: "RegExp()"},
		{Code: "new RegExp('.')"},
		{Code: "new RegExp('abc')"},

		// Valid: with flags
		{Code: "RegExp('.', 'g')"},
		{Code: "new RegExp('.', 'gi')"},
		{Code: "new RegExp('.', 'im')"},
		{Code: "new RegExp('.', 'u')"},
		{Code: "new RegExp('.', 'yu')"},
		{Code: "new RegExp('.', 's')"},
		{Code: "new RegExp('.', 'v')"},

		// Valid: complex patterns
		{Code: "new RegExp('(?<a>b)\\\\k<a>', 'u')"},
		{Code: "new RegExp('(?<=a)b')"},
		{Code: "new RegExp('(?<!a)b')"},

		// Valid: unicode properties
		{Code: "new RegExp('\\\\p{Letter}', 'u')"},

		// Valid: dynamic patterns (can't validate)
		{Code: "RegExp(pattern, 'g')"},
		{Code: "new RegExp('.' + '', 'g')"},
		{Code: "RegExp('{', flags)"},

		// Valid: with allowConstructorFlags option
		{Code: "RegExp('.', 'a')", Options: map[string]interface{}{"allowConstructorFlags": []interface{}{"a"}}},
	}, []rule_tester.InvalidTestCase{
		// Invalid: unterminated character class
		{
			Code: "RegExp('[');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regexMessage",
				},
			},
		},

		// Invalid: unmatched parenthesis
		{
			Code: "new RegExp(')');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regexMessage",
				},
			},
		},

		// Invalid: backslash at end
		{
			Code: "new RegExp('\\\\');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regexMessage",
				},
			},
		},

		// Invalid: invalid flag
		{
			Code: "RegExp('.', 'z');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regexMessage",
				},
			},
		},

		// Invalid: duplicate flags
		{
			Code: "new RegExp('.', 'aa');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regexMessage",
				},
			},
		},

		// Invalid: duplicate flags (different)
		{
			Code: "new RegExp('.', 'gg');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regexMessage",
				},
			},
		},

		// Invalid: conflicting u and v flags
		{
			Code: "new RegExp('.', 'uv');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regexMessage",
				},
			},
		},

		// Invalid: conflicting v and u flags (reversed)
		{
			Code: "new RegExp('.', 'vu');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regexMessage",
				},
			},
		},
	})
}
