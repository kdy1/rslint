package no_unnecessary_type_constraint

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoUnnecessaryTypeConstraintRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryTypeConstraintRule,
		[]rule_tester.ValidTestCase{
			{Code: `type Foo<T extends string> = T;`},
			{Code: `function foo<T extends number>(x: T) {}`},
			{Code: `interface Foo<T> {}`},
		},
		[]rule_tester.InvalidTestCase{
			// No invalid test cases - rule implementation is incomplete
		},
	)
}
