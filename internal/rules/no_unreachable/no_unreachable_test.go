package no_unreachable

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoUnreachableRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnreachableRule,
		[]rule_tester.ValidTestCase{
			{Code: `function foo() { return; }`},
			{Code: `function foo() { throw new Error(); }`},
			{Code: `function foo() { if (true) { return; } else { return; } }`},
			{Code: `function foo() { while (true) { break; } return; }`},
			{Code: `function foo() { for(;;) { break; } return; }`},
			{Code: `function foo() { switch (foo) { case 1: return; } return; }`},
			{Code: `function foo() { try { return; } catch (e) { return; } }`},
			{Code: `function foo() { return; /* comment */ }`},
			{Code: `function foo() { var x = 1; throw new Error(); }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `function foo() { return; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unreachableCode",
						Line:      1,
						Column:    26,
					},
				},
			},
			{
				Code: `function foo() { throw new Error(); var x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unreachableCode",
						Line:      1,
						Column:    37,
					},
				},
			},
			{
				Code: `function foo() { { return; x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unreachableCode",
						Line:      1,
						Column:    28,
					},
				},
			},
			{
				Code: `function foo() { while(true) { break; x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unreachableCode",
						Line:      1,
						Column:    39,
					},
				},
			},
			{
				Code: `function foo() { while(true) { continue; x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unreachableCode",
						Line:      1,
						Column:    42,
					},
				},
			},
			{
				Code: `switch (foo) { case 1: return; var x; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unreachableCode",
						Line:      1,
						Column:    32,
					},
				},
			},
		},
	)
}
