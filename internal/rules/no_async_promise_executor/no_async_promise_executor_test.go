package no_async_promise_executor

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoAsyncPromiseExecutorRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAsyncPromiseExecutorRule,
		[]rule_tester.ValidTestCase{
			{Code: `new Promise((resolve, reject) => {})`},
			{Code: `new Promise((resolve, reject) => {}, async function unrelated() {})`},
			{Code: `new Foo(async (resolve, reject) => {})`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `new Promise(async function foo(resolve, reject) {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "async",
						Line:      1,
						Column:    13,
					},
				},
			},
			{
				Code: `new Promise(async (resolve, reject) => {})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "async",
						Line:      1,
						Column:    13,
					},
				},
			},
			{
				Code: `new Promise(((((async () => {})))))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "async",
						Line:      1,
						Column:    17,
					},
				},
			},
		},
	)
}
