package prefer_template

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferTemplateRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferTemplateRule,
		[]rule_tester.ValidTestCase{
			{Code: `var foo = 'bar';`},
			{Code: `var foo = 'bar' + 'baz';`},
			{Code: `var foo = 'bar' + '\\0';`},
			{Code: "var foo = `bar`;"},
			{Code: "var foo = `hello, ${name}!`;"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `var foo = 'hello, ' + name + '!';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation"},
				},
				Output: []string{"var foo = `hello, ${name}!`;"},
			},
			{
				Code: `var foo = bar + 'baz';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation"},
				},
				Output: []string{"var foo = `${bar}baz`;"},
			},
			{
				Code: `var foo = 'bar' + baz;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation"},
				},
				Output: []string{"var foo = `bar${baz}`;"},
			},
			{
				Code: `var foo = 'bar' + baz + 'qux';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation"},
				},
				Output: []string{"var foo = `bar${baz}qux`;"},
			},
			{
				Code: `var foo = "bar" + baz;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedStringConcatenation"},
				},
				Output: []string{"var foo = `bar${baz}`;"},
			},
		},
	)
}
