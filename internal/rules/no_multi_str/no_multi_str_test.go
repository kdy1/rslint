package no_multi_str

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMultiStrRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoMultiStrRule,
		[]rule_tester.ValidTestCase{
			// Single line string is allowed
			{Code: `var a = 'Line 1 Line 2';`},
		},
		[]rule_tester.InvalidTestCase{
			// Backslash continuation
			{
				Code: "var x = 'Line 1 \\\n Line 2'",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "multilineString",
						Line:      1,
						Column:    9,
					},
				},
			},

			// Function argument with multiline string
			{
				Code: "test('Line 1 \\\n Line 2');",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "multilineString",
						Line:      1,
						Column:    6,
					},
				},
			},

			// Carriage return
			{
				Code: "'foo\\\rbar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "multilineString",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Line separator (U+2028)
			{
				Code: "'foo\\\u2028bar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "multilineString",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Paragraph separator (U+2029)
			{
				Code: "'foo\\\u2029bar';",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "multilineString",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}
