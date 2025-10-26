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
			// String literal modules (ambient modules) are allowed
			{Code: `declare module 'foo';`},
			{Code: `declare module 'foo' {}`},
			// Already using namespace keyword
			{Code: `namespace foo {}`},
			{Code: `declare namespace foo {}`},
			// Global module augmentation
			{Code: `declare global {}`},
		},
		[]rule_tester.InvalidTestCase{
			// Basic module declaration
			{
				Code: `module foo {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useNamespace",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{`namespace foo {}`},
			},
			// Module with declare modifier
			{
				Code: `declare module foo {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useNamespace",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{`declare namespace foo {}`},
			},
			// Multiple nested modules
			{
				Code: `
module Foo {
  module Bar {}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "useNamespace",
						Line:      2,
						Column:    1,
					},
					{
						MessageId: "useNamespace",
						Line:      3,
						Column:    3,
					},
				},
				Output: []string{`
namespace Foo {
  namespace Bar {}
}`},
			},
		},
	)
}
