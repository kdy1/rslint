package no_lone_blocks

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoLoneBlocksRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLoneBlocksRule,
		[]rule_tester.ValidTestCase{
			// Nested control structures
			{Code: `if (foo) { if (bar) { baz(); } }`},
			{Code: `do { bar(); } while (foo)`},
			{Code: `function foo() { while (bar) { baz() } }`},

			// Block-level bindings (let/const)
			{Code: `{ let x = 1; }`},
			{Code: `{ const x = 1; }`},
			{Code: `{ class Bar {} }`},
			{Code: `{ {let y = 1;} let x = 1; }`},

			// Switch cases with blocks
			{Code: `switch (foo) { case 1: { bar(); } break; }`},

			// Function with control flow
			{Code: `function foo() { if (bar) { baz(); } }`},
		},
		[]rule_tester.InvalidTestCase{
			// Empty block
			{
				Code: `{}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantBlock",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Block with only var
			{
				Code: `{var x = 1;}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantBlock",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Standalone block between statements
			{
				Code: `foo(); {} bar();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantBlock",
						Line:      1,
						Column:    8,
					},
				},
			},

			// Nested block within control structure
			{
				Code: `if (foo) { bar(); {} baz(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantNestedBlock",
						Line:      1,
						Column:    19,
					},
				},
			},

			// Nested block in function
			{
				Code: `function foo() { bar(); {} baz(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantNestedBlock",
						Line:      1,
						Column:    25,
					},
				},
			},

			// Nested empty block in loop
			{
				Code: `while (foo) { {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantNestedBlock",
						Line:      1,
						Column:    15,
					},
				},
			},
		},
	)
}
