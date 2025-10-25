package no_unreachable_loop

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoUnreachableLoopRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnreachableLoopRule,
		[]rule_tester.ValidTestCase{
			{Code: `while(foo) { if (bar) { break; } }`},
			{Code: `for(let i = 0; i < 10; i++) { console.log(i); }`},
			{Code: `for(const x of arr) { console.log(x); }`},
			{Code: `for(const key in obj) { console.log(key); }`},
			{Code: `do { console.log('test'); } while (foo);`},
			{Code: `while(true) { if (condition) break; doSomething(); }`},
			{Code: `for(;;) { if (condition) { break; } process(); }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `while(foo) { return; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalid",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `while(foo) { throw new Error(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalid",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `while(foo) { break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalid",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `for(let i = 0; i < 10; i++) { return; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalid",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `for(const x of arr) { break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalid",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `do { return; } while (foo);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "invalid",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}
