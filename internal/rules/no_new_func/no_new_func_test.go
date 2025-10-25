package no_new_func

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewFuncRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNewFuncRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: "var a = new _function(\"b\", \"c\", \"return b+c\");"},
			{Code: "var a = _function(\"b\", \"c\", \"return b+c\");"},
			{Code: "class Function {}; new Function();"},
			{Code: "const fn = () => { class Function {}; new Function(); }"},
			{Code: "function Function() {}; Function();"},
			{Code: "var fn = function () { function Function() {}; Function(); }"},
			{Code: "var x = function Function() { Function(); }"},
			{Code: "call(Function)"},
			{Code: "new Class(Function)"},
			{Code: "foo[Function]()"},
			{Code: "foo(Function.bind)"},
			{Code: "Function.toString()"},
			{Code: "Function[call]()"},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: "var a = new Function(\"b\", \"c\", \"return b+c\");",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 13},
				},
			},
			{
				Code: "var a = Function(\"b\", \"c\", \"return b+c\");",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a = Function.call(null, \"b\", \"c\", \"return b+c\");",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a = Function.apply(null, [\"b\", \"c\", \"return b+c\"]);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a = Function.bind(null, \"b\", \"c\", \"return b+c\")();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a = Function.bind(null, \"b\", \"c\", \"return b+c\");",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a = Function[\"call\"](null, \"b\", \"c\", \"return b+c\");",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 9},
				},
			},
			{
				Code: "var a = (Function?.call)(null, \"b\", \"c\", \"return b+c\");",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 10},
				},
			},
			{
				Code: "const fn = () => { class Function {} }; new Function('', '');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 46},
				},
			},
			{
				Code: "var fn = function () { function Function() {} }; Function('', '');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 51},
				},
			},
		},
	)
}
