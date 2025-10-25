package curly

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCurlyRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &CurlyRule, []rule_tester.ValidTestCase{
		// Valid cases - default "all" mode (braces required)
		{Code: "if (foo) { bar() }"},
		{Code: "if (foo) { bar() } else { baz() }"},
		{Code: "while (foo) { bar() }"},
		{Code: "do { bar() } while (foo)"},
		{Code: "for (;;) { bar() }"},

		// Valid cases - "multi" mode (single statements can omit braces)
		{Code: "if (foo) bar()", Options: []interface{}{"multi"}},
		{Code: "while (foo) bar()", Options: []interface{}{"multi"}},
		{Code: "for (;;) bar()", Options: []interface{}{"multi"}},
		{Code: "if (foo) { bar(); baz(); }", Options: []interface{}{"multi"}},

		// Valid cases - "multi-line" mode
		{Code: "if (foo) bar()", Options: []interface{}{"multi-line"}},
		{Code: "if (foo) { \n bar() \n }", Options: []interface{}{"multi-line"}},

		// Valid cases - "multi-or-nest" mode
		{Code: "if (foo) bar()", Options: []interface{}{"multi-or-nest"}},
		{Code: "if (foo) { if (bar) baz() }", Options: []interface{}{"multi-or-nest"}},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases - default "all" mode
		{
			Code: "if (foo) bar()",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingCurlyAfterCondition"},
			},
			Output: []string{"if (foo) {\nbar()\n}"},
		},
		{
			Code: "while (foo) bar()",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingCurlyAfterCondition"},
			},
			Output: []string{"while (foo) {\nbar()\n}"},
		},
		{
			Code: "if (foo) bar(); else baz()",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingCurlyAfterCondition"},
				{MessageId: "missingCurlyAfter"},
			},
			Output: []string{"if (foo) {\nbar()\n}; else {\nbaz()\n}"},
		},

		// Invalid cases - "multi" mode (unnecessary braces)
		{
			Code:    "if (foo) { bar() }",
			Options: []interface{}{"multi"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedCurlyAfterCondition"},
			},
			Output: []string{"if (foo) bar()"},
		},
		{
			Code:    "while (foo) { bar() }",
			Options: []interface{}{"multi"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedCurlyAfterCondition"},
			},
			Output: []string{"while (foo) bar()"},
		},

		// Invalid cases - "multi-line" mode (missing braces for multi-line)
		{
			Code:    "if (foo) \n bar()",
			Options: []interface{}{"multi-line"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingCurlyAfterCondition"},
			},
			Output: []string{"if (foo) {\n bar()\n}"},
		},

		// Invalid cases - "multi-or-nest" mode (nested requires braces)
		{
			Code:    "if (foo) if (bar) baz()",
			Options: []interface{}{"multi-or-nest"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingCurlyAfterCondition"},
			},
			Output: []string{"if (foo) {\nif (bar) baz()\n}"},
		},

		// Invalid cases - consistent mode
		{
			Code:    "if (foo) { bar() } else baz()",
			Options: []interface{}{"multi", "consistent"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingCurlyAfter"},
			},
			Output: []string{"if (foo) { bar() } else {\nbaz()\n}"},
		},
	})
}
