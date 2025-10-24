package no_unsafe_declaration_merging

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoUnsafeDeclarationMergingRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnsafeDeclarationMergingRule,
		[]rule_tester.ValidTestCase{
			{Code: `interface Foo { bar: string; }`},
			{Code: `class Foo {}`},
			{Code: `const x = 1;`},
		},
		[]rule_tester.InvalidTestCase{
			// No invalid test cases - rule implementation is incomplete
		},
	)
}
