package prefer_namespace_keyword

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferNamespaceKeywordRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNamespaceKeywordRule, []rule_tester.ValidTestCase{
		// Valid: declare module with string literal (external API)
		{Code: `declare module 'foo';`},
		// Valid: declare module with string literal and body
		{Code: `declare module 'foo' {}`},
		// Valid: namespace keyword
		{Code: `namespace foo {}`},
		// Valid: declare namespace
		{Code: `declare namespace foo {}`},
		// Valid: declare global
		{Code: `declare global {}`},
	}, []rule_tester.InvalidTestCase{
		// Invalid: module keyword without string literal
		{
			Code:   `module foo {}`,
			Output: []string{`namespace foo {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 14,
				},
			},
		},
		// Invalid: declare module with identifier (not string literal)
		{
			Code:   `declare module foo {}`,
			Output: []string{`declare namespace foo {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 22,
				},
			},
		},
		// Invalid: nested module declarations
		{
			Code: `declare module foo {
  declare module bar {}
}`,
			Output: []string{`declare namespace foo {
  declare namespace bar {}
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
					Line:      1,
					Column:    1,
					EndLine:   3,
					EndColumn: 2,
				},
				{
					MessageId: "useNamespace",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 24,
				},
			},
		},
	})
}
