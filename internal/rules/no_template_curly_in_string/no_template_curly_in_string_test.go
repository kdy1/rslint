package no_template_curly_in_string

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoTemplateCurlyInStringRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoTemplateCurlyInStringRule,
		[]rule_tester.ValidTestCase{
			{Code: "const str = `Hello, ${name}`;"},
			{Code: "const str = `Hello, name`;"},
			{Code: "const str = 'Hello, name';"},
			{Code: "const str = 'Hello, ' + name;"},
			{Code: "const str = '$2';"},
			{Code: "const str = '${';"},
			{Code: "const str = '$}';"},
			{Code: "const str = '{foo}';"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "const str = 'Hello, ${name}';",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedTemplateExpression",
						Line:      1,
						Column:    13,
					},
				},
			},
			{
				Code: `const str = "Hello, ${name}";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedTemplateExpression",
						Line:      1,
						Column:    13,
					},
				},
			},
			{
				Code: `const str = '${greeting}, ${name}';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedTemplateExpression",
						Line:      1,
						Column:    13,
					},
				},
			},
			{
				Code: `const str = 'Hello, ${index + 1}';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedTemplateExpression",
						Line:      1,
						Column:    13,
					},
				},
			},
		},
	)
}
