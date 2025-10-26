package no_alert

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAlertRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAlertRule,
		[]rule_tester.ValidTestCase{
			// Bracket notation with non-alert identifier
			{Code: `a[o.k](1)`},

			// alert on custom object
			{Code: `foo.alert(foo)`},
			{Code: `foo.confirm(foo)`},
			{Code: `foo.prompt(foo)`},

			// Locally defined alert function
			{Code: `function alert() {} alert();`},

			// alert as local variable
			{Code: `var alert = function() {}; alert();`},

			// Scoped alert variable
			{Code: `function foo() { var alert = bar; alert(); }`},

			// alert as function parameter
			{Code: `function foo(alert) { alert(); }`},

			// Separate scope alert
			{Code: `var alert = function() {}; function test() { alert(); }`},

			// Nested scoped alert
			{Code: `function foo() { var alert = function() {}; function test() { alert(); } }`},

			// Locally defined confirm
			{Code: `function confirm() {} confirm();`},

			// Locally defined prompt
			{Code: `function prompt() {} prompt();`},

			// Dynamic window property access
			{Code: `window[alert]();`},

			// alert on this context
			{Code: `function foo() { this.alert(); }`},

			// Shadowed window variable
			{Code: `function foo() { var window = bar; window.alert(); }`},

			// globalThis reference
			{Code: `globalThis.alert();`, LanguageOptions: &rule_tester.LanguageOptions{
				EcmaVersion: rule_tester.IntPtr(5),
			}},
			{Code: `globalThis['alert']();`, LanguageOptions: &rule_tester.LanguageOptions{
				EcmaVersion: rule_tester.IntPtr(6),
			}},
			{Code: `globalThis.alert();`, LanguageOptions: &rule_tester.LanguageOptions{
				EcmaVersion: rule_tester.IntPtr(2017),
			}},
		},
		[]rule_tester.InvalidTestCase{
			// Bare alert call
			{
				Code: `alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Direct window.alert
			{
				Code: `window.alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Computed window property
			{
				Code: `window['alert'](foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Bare confirm call
			{
				Code: `confirm(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-confirm",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Direct window.confirm
			{
				Code: `window.confirm(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-confirm",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Computed window confirm
			{
				Code: `window['confirm'](foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-confirm",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Bare prompt call
			{
				Code: `prompt(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-prompt",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Direct window.prompt
			{
				Code: `window.prompt(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-prompt",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Computed window prompt
			{
				Code: `window['prompt'](foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-prompt",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Local definition shadowed by window call
			{
				Code: `function alert() {} window.alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    21,
					},
				},
			},

			// Multi-line with local variable
			{
				Code: `var alert = function() {};
window.alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      2,
						Column:    1,
					},
				},
			},

			// Parameter shadowed by window
			{
				Code: `function foo(alert) { window.alert(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    23,
					},
				},
			},

			// Unscoped alert call
			{
				Code: `function foo() { alert(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    18,
					},
				},
			},

			// Multi-line with scoped variable
			{
				Code: `function foo() { var alert = function() {}; }
alert();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      2,
						Column:    1,
					},
				},
			},

			// alert on this object
			{
				Code: `this.alert(foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Computed alert on this
			{
				Code: `this['alert'](foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    1,
					},
				},
			},

			// globalThis.alert with ES2020
			{
				Code: `globalThis.alert();`,
				LanguageOptions: &rule_tester.LanguageOptions{
					EcmaVersion: rule_tester.IntPtr(2020),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    1,
					},
				},
			},

			// globalThis['alert'] with ES2020
			{
				Code: `globalThis['alert'](foo)`,
				LanguageOptions: &rule_tester.LanguageOptions{
					EcmaVersion: rule_tester.IntPtr(2020),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Optional chaining: window?.alert
			{
				Code: `window?.alert(foo)`,
				LanguageOptions: &rule_tester.LanguageOptions{
					EcmaVersion: rule_tester.IntPtr(2020),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    1,
					},
				},
			},

			// Optional chaining with parentheses
			{
				Code: `(window?.alert)(foo)`,
				LanguageOptions: &rule_tester.LanguageOptions{
					EcmaVersion: rule_tester.IntPtr(2020),
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected-alert",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}
