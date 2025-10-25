package no_obj_calls

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoObjCallsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoObjCallsRule,
		[]rule_tester.ValidTestCase{
			// Direct method calls (allowed)
			{Code: "var x = Math;"},
			{Code: "var x = Math.random();"},
			{Code: "var x = Math.PI;"},
			{Code: "JSON.parse(foo)"},

			// Namespaced calls (allowed)
			{Code: "var x = foo.Math();"},
			{Code: "var x = new foo.Math();"},
			{Code: "var x = new Math.foo;"},

			// Global object access with globalThis (ES2020+)
			{Code: "globalThis.Math();", FileName: "test.js"},
			{Code: "var x = globalThis.JSON();", FileName: "test.js"},
		},
		[]rule_tester.InvalidTestCase{
			// Direct calls on non-callable global objects
			{
				Code: "Math();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: "var x = JSON();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: "new Math();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: "Reflect();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: "Atomics();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: "Intl();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},

			// With globalThis (ES2020+)
			{
				Code:     "var x = globalThis.Math();",
				FileName: "test.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code:     "var x = new globalThis.JSON();",
				FileName: "test.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
		},
	)
}
