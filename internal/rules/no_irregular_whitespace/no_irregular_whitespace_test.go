package no_irregular_whitespace

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoIrregularWhitespaceRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoIrregularWhitespaceRule, []rule_tester.ValidTestCase{
		// Valid: regular whitespace
		{Code: "var foo = 'bar';"},
		{Code: "var foo = 'bar\\t';"},
		{Code: "var foo = 'bar\\n';"},
		{Code: "var foo = 'bar\\r\\n';"},

		// Valid: escaped unicode sequences
		{Code: "var foo = '\\u000B';"},
		{Code: "var foo = '\\u000C';"},
		{Code: "var foo = '\\u0085';"},
		{Code: "var foo = '\\u00A0';"},
		{Code: "var foo = '\\u180E';"},
		{Code: "var foo = '\\uFEFF';"},
		{Code: "var foo = '\\u2000';"},
		{Code: "var foo = '\\u2001';"},
		{Code: "var foo = '\\u2028';"},
		{Code: "var foo = '\\u2029';"},
		{Code: "var foo = '\\u3000';"},

		// Valid: irregular whitespace in strings (default skipStrings: true)
		{Code: "var foo = '\u00A0';"},
		{Code: "var foo = '\u2000';"},
		{Code: "var foo = 'test\u00A0test';"},

		// Valid: with skipComments option
		{Code: "// comment\u00A0", Options: map[string]interface{}{"skipComments": true}},

		// Valid: with skipTemplates option
		{Code: "`template\u00A0`", Options: map[string]interface{}{"skipTemplates": true}},

		// Valid: with skipRegExps option
		{Code: "/regexp\u00A0/", Options: map[string]interface{}{"skipRegExps": true}},
	}, []rule_tester.InvalidTestCase{
		// Invalid: irregular whitespace in variable names/code
		// Note: These tests would need actual irregular whitespace in identifiers,
		// which is typically prevented by the parser. Instead, we test strings
		// when skipStrings is explicitly set to false

		// Invalid: irregular whitespace in strings when skipStrings: false
		{
			Code:    "var foo = '\u00A0';",
			Options: map[string]interface{}{"skipStrings": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIrregularWhitespace",
				},
			},
		},

		// Invalid: irregular whitespace in templates when skipTemplates: false
		{
			Code:    "`template\u00A0`",
			Options: map[string]interface{}{"skipTemplates": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIrregularWhitespace",
				},
			},
		},

		// Invalid: multiple irregular whitespace types
		{
			Code:    "var foo = '\u2000\u2001\u2002';",
			Options: map[string]interface{}{"skipStrings": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIrregularWhitespace",
				},
			},
		},

		// Invalid: BOM character
		{
			Code:    "var foo = '\uFEFF';",
			Options: map[string]interface{}{"skipStrings": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIrregularWhitespace",
				},
			},
		},

		// Invalid: line separator
		{
			Code:    "var foo = '\u2028';",
			Options: map[string]interface{}{"skipStrings": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIrregularWhitespace",
				},
			},
		},

		// Invalid: paragraph separator
		{
			Code:    "var foo = '\u2029';",
			Options: map[string]interface{}{"skipStrings": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIrregularWhitespace",
				},
			},
		},

		// Invalid: ideographic space
		{
			Code:    "var foo = '\u3000';",
			Options: map[string]interface{}{"skipStrings": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noIrregularWhitespace",
				},
			},
		},
	})
}
