package no_unexpected_multiline

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoUnexpectedMultilineRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnexpectedMultilineRule,
		[]rule_tester.ValidTestCase{
			// Valid cases - no unexpected multiline
			{Code: "(x || y).aFunction()"},
			{Code: "[a, b, c].forEach(doSomething)"},
			{Code: "var a = b;\n(x || y).doSomething()"},
			{Code: "var a = b\n;(x || y).doSomething()"},
			{Code: "var a = b\nvoid (x || y).doSomething()"},
			{Code: "var a = b;\n[1, 2, 3].forEach(console.log)"},
			{Code: "var a = b\nvoid [1, 2, 3].forEach(console.log)"},
			{Code: "\"abc\\\n(123)\""},
			{Code: "var a = (\n(123)\n)"},
			{Code: "f(\n(x)\n)"},
			{Code: "(\nfunction () {}\n)[1]"},

			// Template literals
			{Code: "foo\n`bar`"},
			{Code: "foo\n  `bar`"},
			{Code: "foo()\n`bar`"},

			// Division operators
			{Code: "x\n/foo/"},
			{Code: "x\n/foo/g"},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid cases - unexpected multiline
			{
				Code: "var a = b\n(x || y).doSomething()",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function"},
				},
			},
			{
				Code: "var a = (a || b)\n(x || y).doSomething()",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function"},
				},
			},
			{
				Code: "var a = (a || b)\n(x).doSomething()",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function"},
				},
			},
			{
				Code: "var a = b\n[a, b, c].forEach(doSomething)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "property"},
				},
			},
			{
				Code: "var a = b\n    (x || y).doSomething()",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function"},
				},
			},
			{
				Code: "var a = b\n  [a, b, c].forEach(doSomething)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "property"},
				},
			},

			// Tagged templates
			{
				Code: "foo\n`bar`;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "taggedTemplate"},
				},
			},
		},
	)
}
