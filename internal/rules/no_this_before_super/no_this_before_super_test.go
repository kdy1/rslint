package no_this_before_super

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoThisBeforeSuperRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoThisBeforeSuperRule,
		[]rule_tester.ValidTestCase{
			{Code: `class A { constructor() { } }`},
			{Code: `class A extends B { constructor() { super(); this.foo = 1; } }`},
			{Code: `class A extends B { constructor() { super(); this.foo(); } }`},
			{Code: `class A extends B { foo() { this.bar = 1; } }`},
			{Code: `class A extends null { constructor() { this.foo = 1; } }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `class A extends B { constructor() { this.foo = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noBeforeSuper",
						Line:      1,
						Column:    37,
					},
				},
			},
			{
				Code: `class A extends B { constructor() { this.foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noBeforeSuper",
						Line:      1,
						Column:    37,
					},
				},
			},
			{
				Code: `class A extends B { constructor() { super.foo(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noBeforeSuper",
						Line:      1,
						Column:    37,
					},
				},
			},
		},
	)
}
