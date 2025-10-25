package no_setter_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSetterReturnRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoSetterReturnRule,
		[]rule_tester.ValidTestCase{
			// Regular functions (not setters)
			{Code: "function foo() { return 1; }"},
			{Code: "function set(val) { return 1; }"},
			{Code: "var foo = function() { return 1; };"},
			{Code: "var foo = function set() { return 1; };"},

			// Setters without return values
			{Code: "({ set foo(val) { return; } });"},
			{Code: "class A { set foo(val) { return; } }"},
			{Code: "({ set foo(val) { if (val) { return; } } });"},

			// Getters (allowed to return)
			{Code: "({ get foo() { return 1; } });"},
			{Code: "class A { get foo() { return 1; } }"},

			// Regular methods
			{Code: "({ set(val) { return 1; } });"},
			{Code: "class A { set(val) { return 1; } }"},
			{Code: "class A { constructor(val) { return 1; } }"},

			// Object.defineProperty with no return
			{Code: "Object.defineProperty(foo, 'bar', { set(val) { return; } });"},
			{Code: "Object.defineProperty(foo, 'bar', { set(val) { if (val) { return; } } });"},
		},
		[]rule_tester.InvalidTestCase{
			// Basic setter violations
			{
				Code: "({ set a(val) { return val; } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},
			{
				Code: "({ set a(val) { return val + 1; } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},
			{
				Code: "class A { set a(val) { return 1; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},
			{
				Code: "class A { static set a(val) { return 1; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},

			// Various return types
			{
				Code: "({ set a(val) { return this.b; } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},
			{
				Code: "({ set a(val) { return undefined; } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},
			{
				Code: "({ set a(val) { return null; } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},

			// Multiple returns
			{
				Code: "({ set a(val) { if (val) { return 1; } else { return 2; } } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
					{MessageId: "returnsValue"},
				},
			},

			// Multiple setters
			{
				Code: "({ set a(val) { return 1; }, set b(val) { return 1; } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
					{MessageId: "returnsValue"},
				},
			},

			// Object.defineProperty
			{
				Code: "Object.defineProperty(foo, 'bar', { set(val) { return 1; } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},
			{
				Code: "Object.defineProperty(foo, 'bar', { set: function(val) { return 1; } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},

			// Object.defineProperties
			{
				Code: "Object.defineProperties(foo, { bar: { set(val) { return 1; } } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},

			// Object.create
			{
				Code: "Object.create(null, { bar: { set(val) { return 1; } } });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},

			// Arrow function implicit return
			{
				Code: "Object.defineProperty(foo, 'bar', { set: val => val });",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "returnsValue"},
				},
			},
		})
}
