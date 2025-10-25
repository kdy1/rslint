package no_implied_eval

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoImpliedEvalRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoImpliedEvalRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: "setTimeout;"},
			{Code: "setTimeout = foo;"},
			{Code: "window.setTimeout;"},
			{Code: "setTimeout();"},
			{Code: "window.setTimeout(function() { x = 1; }, 100);"},
			{Code: "setTimeout(foo, 10);"},
			{Code: "setTimeout(foo);"},
			{Code: "setInterval(foo, 10);"},
			{Code: "var window; window.setTimeout('foo', 100);"},
			{Code: "setTimeout(function() { alert('Hi!'); }, 10);"},
			{Code: "setInterval(function() { alert('Hi!'); }, 10);"},
			{Code: "execScript(function() { alert('Hi!'); });"},
			{Code: "window.execScript(function() { alert('Hi!'); });"},
			{Code: "setTimeout(foo());"},
			{Code: "setTimeout(() => {});"},
			{Code: "setInterval(() => {});"},
			{Code: "setTimeout(`foo${bar}`);"},  // Template with variables could be safe
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: "setTimeout(\"x = 1;\");",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "setTimeout('x = 1;');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "setInterval('x = 1;');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "window.setTimeout('x = 1;');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 8},
				},
			},
			{
				Code: "window.setInterval('x = 1;');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 8},
				},
			},
			{
				Code: "global.setTimeout('x = 1;');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 8},
				},
			},
			{
				Code: "globalThis.setTimeout('x = 1;');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 12},
				},
			},
			{
				Code: "setTimeout('foo' + bar);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "setTimeout(`foo`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "execScript('x = 1;');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: "window.execScript('x = 1;');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 8},
				},
			},
			{
				Code: "execScript(foo);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
		},
	)
}
