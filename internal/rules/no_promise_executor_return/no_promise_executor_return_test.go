package no_promise_executor_return

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoPromiseExecutorReturnRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoPromiseExecutorReturnRule,
		[]rule_tester.ValidTestCase{
			// Empty return (control flow) is allowed
			{Code: "new Promise((resolve, reject) => { if (x) { return; } resolve(); })"},
			{Code: "new Promise(function (resolve, reject) { if (x) { return; } resolve(); })"},

			// No return statement
			{Code: "new Promise((resolve, reject) => { resolve(1); })"},
			{Code: "new Promise((r) => { r(1); })"},

			// Nested functions can return
			{Code: "new Promise(() => { function foo() { return 1; } })"},
			{Code: "new Promise(() => { const foo = () => { return 1; }; })"},

			// Not Promise constructor
			{Code: "Promise()"},
			{Code: "new Foo((resolve) => resolve(1))"},
		},
		[]rule_tester.InvalidTestCase{
			// Basic violations - arrow functions with expression body
			{
				Code: "new Promise((resolve, reject) => resolve(1))",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 38},
				},
			},
			{
				Code: "new Promise((resolve, reject) => 1)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 34},
				},
			},

			// Block body with return statement
			{
				Code: "new Promise(function (resolve, reject) { return 1; })",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 42},
				},
			},
			{
				Code: "new Promise((resolve, reject) => { return 1; })",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 36},
				},
			},
			{
				Code: "new Promise((resolve, reject) => { if (x) { return 1; } })",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 45},
				},
			},
		},
	)
}

func TestNoPromiseExecutorReturnRuleWithOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoPromiseExecutorReturnRule,
		[]rule_tester.ValidTestCase{
			// With allowVoid: true
			{
				Code:    "new Promise((r) => void r(1))",
				Options: map[string]interface{}{"allowVoid": true},
			},
			{
				Code:    "new Promise((r) => { return void r(1); })",
				Options: map[string]interface{}{"allowVoid": true},
			},
			{
				Code:    "new Promise(r => void 0)",
				Options: map[string]interface{}{"allowVoid": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			// With allowVoid: false (default), void expressions are not allowed
			{
				Code: "new Promise((r) => void r(1))",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue", Line: 1, Column: 24},
				},
				Options: map[string]interface{}{"allowVoid": false},
			},
		},
	)
}
