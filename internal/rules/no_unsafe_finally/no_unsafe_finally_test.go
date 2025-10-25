package no_unsafe_finally

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoUnsafeFinallyRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnsafeFinallyRule,
		[]rule_tester.ValidTestCase{
			{Code: `try { foo(); } catch(e) { bar(); } finally { console.log('done'); }`},
			{Code: `try { foo(); } finally { var x = 1; }`},
			{Code: `try { foo(); } finally { function bar() { return 1; } }`},
			{Code: `try { foo(); } finally { const fn = () => { return 1; }; }`},
			{Code: `try { foo(); } finally { while(true) { break; } }`},
			{Code: `try { foo(); } finally { switch(x) { case 1: break; } }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `try { foo(); } finally { return 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeUsage",
						Line:      1,
						Column:    26,
					},
				},
			},
			{
				Code: `try { foo(); } finally { throw new Error(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeUsage",
						Line:      1,
						Column:    26,
					},
				},
			},
			{
				Code: `while (true) { try { foo(); } finally { break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeUsage",
						Line:      1,
						Column:    41,
					},
				},
			},
			{
				Code: `while (true) { try { foo(); } finally { continue; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeUsage",
						Line:      1,
						Column:    41,
					},
				},
			},
			{
				Code: `try { foo(); } finally { if (bar) { return; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeUsage",
						Line:      1,
						Column:    37,
					},
				},
			},
		},
	)
}
