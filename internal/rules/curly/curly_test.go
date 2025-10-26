package curly

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCurlyRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&CurlyRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Default "all" mode - requires braces everywhere
			{Code: `if (foo) { bar() }`},
			{Code: `if (foo) { bar() } else { baz() }`},
			{Code: `while (foo) { bar() }`},
			{Code: `do { bar(); } while (foo)`},
			{Code: `for (;;) { bar() }`},
			{Code: `for (var foo in bar) { console.log(foo) }`},
			{Code: `for (var foo of bar) { console.log(foo) }`},

			// Multi mode - allows single statements without braces
			{
				Code:    `if (foo) bar()`,
				Options: []interface{}{"multi"},
			},
			{
				Code:    `if (foo) bar(); else baz()`,
				Options: []interface{}{"multi"},
			},
			{
				Code:    `while (foo) bar()`,
				Options: []interface{}{"multi"},
			},
			{
				Code:    `do bar(); while (foo)`,
				Options: []interface{}{"multi"},
			},
			{
				Code:    `for (;;) bar()`,
				Options: []interface{}{"multi"},
			},
			{
				Code:    `for (var foo in bar) console.log(foo)`,
				Options: []interface{}{"multi"},
			},
			{
				Code:    `for (var foo of bar) console.log(foo)`,
				Options: []interface{}{"multi"},
			},
			// Multi mode - requires braces for multi-statement blocks
			{
				Code:    `if (a) { b; c; }`,
				Options: []interface{}{"multi"},
			},
			{
				Code:    `while (a) { b; c; }`,
				Options: []interface{}{"multi"},
			},

			// Multi-line mode - allows single-line statements without braces
			{
				Code:    `if (foo) bar()`,
				Options: []interface{}{"multi-line"},
			},
			{
				Code:    `while (foo) bar()`,
				Options: []interface{}{"multi-line"},
			},
			{
				Code: `if (foo) {
					bar()
				}`,
				Options: []interface{}{"multi-line"},
			},

			// Multi-or-nest mode
			{
				Code:    `if (foo) bar()`,
				Options: []interface{}{"multi-or-nest"},
			},
			{
				Code: `if (foo) {
					bar()
				}`,
				Options: []interface{}{"multi-or-nest"},
			},
			{
				Code:    `if (a) { if (b) c(); }`,
				Options: []interface{}{"multi-or-nest"},
			},

			// Consistent mode
			{
				Code:    `if (foo) { bar() } else { baz() }`,
				Options: []interface{}{"multi", "consistent"},
			},
			{
				Code:    `if (foo) bar(); else baz()`,
				Options: []interface{}{"multi", "consistent"},
			},

			// ES6 features
			{
				Code:    `for (const foo of bar) { console.log(foo) }`,
				Options: []interface{}{"all"},
			},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Missing braces with "all" mode (default)
			{
				Code: `if (foo) bar()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
				},
			},
			{
				Code: `while (foo) bar()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
				},
			},
			{
				Code: `do bar(); while (foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
				},
			},
			{
				Code: `for (;;) bar()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
				},
			},
			{
				Code: `for (var foo in bar) console.log(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
				},
			},
			{
				Code: `for (var foo of bar) console.log(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
				},
			},
			{
				Code: `if (foo) bar(); else baz()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
					{MessageId: "missingCurlyAfterElse"},
				},
			},

			// Unnecessary braces with "multi" mode
			{
				Code:    `if (foo) { bar() }`,
				Options: []interface{}{"multi"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter"},
				},
			},
			{
				Code:    `while (foo) { bar() }`,
				Options: []interface{}{"multi"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter"},
				},
			},
			{
				Code:    `for (;;) { bar() }`,
				Options: []interface{}{"multi"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter"},
				},
			},
			{
				Code:    `if (foo) { bar() } else { baz() }`,
				Options: []interface{}{"multi"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter"},
					{MessageId: "unexpectedCurlyAfterElse"},
				},
			},

			// Multi-line violations
			{
				Code: `if (foo)
					bar()`,
				Options: []interface{}{"multi-line"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
				},
			},
			{
				Code: `while (foo)
					bar()`,
				Options: []interface{}{"multi-line"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
				},
			},

			// Multi-or-nest violations
			{
				Code: `if (foo)
					bar()`,
				Options: []interface{}{"multi-or-nest"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
				},
			},
			{
				Code:    `if (a) if (b) c()`,
				Options: []interface{}{"multi-or-nest"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
				},
			},

			// Consistent violations
			{
				Code:    `if (foo) { bar() } else baz()`,
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "inconsistentCurly"},
				},
			},
			{
				Code:    `if (foo) bar(); else { baz() }`,
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "inconsistentCurly"},
				},
			},

			// Nested if-else
			{
				Code: `if (foo) bar(); else if (baz) qux()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter"},
					{MessageId: "missingCurlyAfter"},
				},
			},
		},
	)
}
