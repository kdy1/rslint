package no_case_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCaseDeclarationsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoCaseDeclarationsRule,
		[]rule_tester.ValidTestCase{
			// Valid: let declaration wrapped in braces
			{Code: `switch (a) { case 1: { let x = 1; break; } default: { let x = 2; break; } }`},

			// Valid: const declaration wrapped in braces
			{Code: `switch (a) { case 1: { const x = 1; break; } default: { const x = 2; break; } }`},

			// Valid: function declaration wrapped in braces
			{Code: `switch (a) { case 1: { function f() {} break; } default: { function f() {} break; } }`},

			// Valid: class declaration wrapped in braces
			{Code: `switch (a) { case 1: { class C {} break; } default: { class C {} break; } }`},

			// Valid: multiple case labels without declarations
			{Code: `switch (a) { case 1: case 2: let x = 1; break; }`},

			// Valid: var declaration (no braces needed)
			{Code: `switch (a) { case 1: var x = 1; break; }`},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid: function declaration without braces in case
			{
				Code: `switch (a) { case 1: function f() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    22,
					},
				},
			},

			// Invalid: function declaration without braces in default
			{
				Code: `switch (a) { default: function f() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    23,
					},
				},
			},

			// Invalid: class declaration without braces in case
			{
				Code: `switch (a) { case 1: class C {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    22,
					},
				},
			},

			// Invalid: class declaration without braces in default
			{
				Code: `switch (a) { default: class C {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    23,
					},
				},
			},

			// Invalid: single let without braces
			{
				Code: `switch (a) { case 1: let x = 1; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    22,
					},
				},
			},

			// Invalid: multiple cases with let (2 errors)
			{
				Code: `switch (a) { case 1: let x = 1; break; case 2: let y = 2; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    22,
					},
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    49,
					},
				},
			},

			// Invalid: case and default with let (2 errors)
			{
				Code: `switch (a) { case 1: let x = 1; break; default: let y = 2; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    22,
					},
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    50,
					},
				},
			},

			// Invalid: const without braces in case
			{
				Code: `switch (a) { case 1: const x = 1; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    22,
					},
				},
			},

			// Invalid: const without braces in default
			{
				Code: `switch (a) { default: const x = 1; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    23,
					},
				},
			},

			// Invalid: function declaration in case with default clause
			{
				Code: `switch (a) { case 1: function f() {} break; default: f(); break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    22,
					},
				},
			},
		},
	)
}
