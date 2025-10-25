package no_undef

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoUndefRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUndefRule,
		[]rule_tester.ValidTestCase{
			// Basic valid cases
			{Code: "var a = 1, b = 2; a;"},
			{Code: "function a(){} a();"},
			{Code: "function f(b) { b; }"},
			{Code: "var a; a = 1; a++;"},
			{Code: "var a; function f() { a = 1; }"},

			// typeof expressions (default behavior)
			{Code: "typeof a"},
			{Code: "typeof (a)"},
			{Code: "var b = typeof a"},
			{Code: "typeof a === 'undefined'"},
			{Code: "if (typeof a === 'undefined') {}"},

			// Object and built-in globals
			{Code: "Object; isNaN();"},
			{Code: "toString()"},

			// ES6 features
			{Code: "function foo() { var [a, b=4] = [1, 2]; return {a, b}; }"},
			{Code: "var toString = 1;"},
			{Code: "function myFunc(...foo) { return foo;}"},
			{Code: "var console; [1,2,3].forEach(obj => { console.log(obj); });"},

			// Destructuring
			{Code: "var a; [a] = [0];"},
			{Code: "var a; ({a} = {});"},
			{Code: "var a; ({b: a} = {});"},

			// Class extends
			{Code: "var Foo; class Bar extends Foo { constructor() { super(); }}"},
		},
		[]rule_tester.InvalidTestCase{
			// Basic invalid cases
			{
				Code: "a = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef"},
				},
			},
			{
				Code: "var a = b;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef"},
				},
			},
			{
				Code: "function f() { b; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef"},
				},
			},

			// Destructuring with undefined vars
			{
				Code: "[a] = [0];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef"},
				},
			},
			{
				Code: "({a} = {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef"},
				},
			},
			{
				Code: "({b: a} = {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef"},
				},
			},
		},
	)
}

func TestNoUndefRuleWithTypeofOption(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUndefRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			// With typeof: true option, check typeof operands
			{
				Code: "if (typeof anUndefinedVar === 'string') {}",
				Options: map[string]interface{}{
					"typeof": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "undef"},
				},
			},
		},
	)
}
