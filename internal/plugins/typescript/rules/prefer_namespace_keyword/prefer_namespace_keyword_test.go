package prefer_namespace_keyword

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestPreferNamespaceKeywordRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferNamespaceKeywordRule,
		[]rule_tester.ValidTestCase{
			// External module declaration (valid use of module)
			{Code: `declare module 'foo';`},
			{Code: `declare module 'foo' {}`},
			// Already using namespace
			{Code: `namespace foo {}`},
			{Code: `declare namespace foo {}`},
			// Global declaration
			{Code: `declare global {}`},
		},
		[]rule_tester.InvalidTestCase{
			// Using module instead of namespace
			{
				Code: `module foo {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useNamespace",
					},
				},
			},
			// Using module with declare
			{
				Code: `declare module foo {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useNamespace",
					},
				},
			},
		},
	)
}
