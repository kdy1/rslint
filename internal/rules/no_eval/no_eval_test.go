package no_eval

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEvalRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoEvalRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: "Eval(foo)"},
			{Code: "setTimeout('foo')"},
			{Code: "setInterval('foo')"},
			{Code: "window.setTimeout('foo')"},
			{Code: "window.setInterval('foo')"},
			{Code: "window.noeval('foo')"},
			{Code: "function foo() { var eval = 'foo'; window[eval]('foo') }"},
			{Code: "global.noeval('foo')"},
			{Code: "globalThis.noneval('foo')"},
			{Code: "this.noeval('foo');"},
			{Code: "function foo() { 'use strict'; this.eval('foo'); }"},
			{Code: "var obj = {foo: function() { this.eval('foo'); }}"},
			{Code: "class A { foo() { this.eval(); } }"},
			{Code: "class A { static foo() { this.eval(); } }"},
			// With allowIndirect option
			{Code: "(0, eval)('foo')", Options: map[string]interface{}{"allowIndirect": true}},
			{Code: "(0, window.eval)('foo')", Options: map[string]interface{}{"allowIndirect": true}},
			{Code: "var EVAL = eval; EVAL('foo')", Options: map[string]interface{}{"allowIndirect": true}},
			{Code: "window.eval('foo')", Options: map[string]interface{}{"allowIndirect": true}},
			{Code: "this.eval('foo')", Options: map[string]interface{}{"allowIndirect": true}},
			{Code: "global.eval('foo')", Options: map[string]interface{}{"allowIndirect": true}},
			{Code: "globalThis.eval('foo')", Options: map[string]interface{}{"allowIndirect": true}},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: "eval(foo)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: "eval('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: "function foo(eval) { eval('foo') }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			{
				Code: "(0, eval)('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5},
				},
			},
			{
				Code: "var EVAL = eval; EVAL('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			{
				Code: "window.eval('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
				},
			},
			{
				Code: "this.eval('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			{
				Code: "globalThis.eval('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			{
				Code: "global.eval('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
				},
			},
		},
	)
}
