package no_console

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConsoleRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoConsoleRule, []rule_tester.ValidTestCase{
		// Valid cases - no console usage
		{Code: "Console.log(foo)"},
		{Code: "console"},
		{Code: "var console = require('myconsole'); console.log(foo)"},

		// Valid cases - with allow option
		{Code: "console.info(foo)", Options: []interface{}{map[string]interface{}{"allow": []interface{}{"info"}}}},
		{Code: "console.warn(foo)", Options: []interface{}{map[string]interface{}{"allow": []interface{}{"warn"}}}},
		{Code: "console.error(foo)", Options: []interface{}{map[string]interface{}{"allow": []interface{}{"error"}}}},
		{Code: "console.log(foo)", Options: []interface{}{map[string]interface{}{"allow": []interface{}{"log"}}}},
		{Code: "console.warn(foo); console.info(foo)", Options: []interface{}{map[string]interface{}{"allow": []interface{}{"warn", "info"}}}},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases - no options
		{
			Code: "console.log(foo)",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
		},
		{
			Code: "console.error(foo)",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
		},
		{
			Code: "console.info(foo)",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
		},
		{
			Code: "console.warn(foo)",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
		},

		// Invalid cases - with allow option (methods not in allow list)
		{
			Code:    "console.log(foo)",
			Options: []interface{}{map[string]interface{}{"allow": []interface{}{"warn"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
		},
		{
			Code:    "console.error(foo)",
			Options: []interface{}{map[string]interface{}{"allow": []interface{}{"warn"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
		},

		// Invalid cases - bracket notation
		{
			Code: "console['log'](foo)",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
		},

		// Invalid cases - in different contexts
		{
			Code: "if (a) console.log(foo)",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
		},
		{
			Code: "function bar() { console.log(foo) }",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpected"},
			},
		},
	})
}
