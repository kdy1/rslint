package no_inner_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInnerDeclarationsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInnerDeclarationsRule, []rule_tester.ValidTestCase{
		// Valid: top-level function declarations
		{Code: "function doSomething() { }"},
		{Code: "function doSomethingElse() { }"},

		// Valid: function declarations in function bodies
		{Code: "function outer() { function inner() { } }"},
		{Code: "(function() { function inner() { } }());"},

		// Valid: function expressions (not declarations)
		{Code: "if (test) { var foo = function() { }; }"},
		{Code: "if (test) { foo = function() { }; }"},

		// Valid: var declarations with default "functions" mode
		{Code: "if (test) { var foo = 42; }"},

		// Valid: let/const declarations (always allowed)
		{Code: "if (test) { let foo = 42; }"},
		{Code: "if (test) { const foo = 42; }"},

		// Valid: class methods
		{Code: "class C { method() { function inner() { } } }"},
	}, []rule_tester.InvalidTestCase{
		// Invalid: function declaration in if block
		{
			Code: "if (test) { function doSomething() { } }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moveDeclToRoot",
				},
			},
		},

		// Invalid: function declaration in for loop
		{
			Code: "for (let i = 0; i < 10; i++) { function doSomething() { } }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moveDeclToRoot",
				},
			},
		},

		// Invalid: function declaration in while loop
		{
			Code: "while (test) { function doSomething() { } }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moveDeclToRoot",
				},
			},
		},

		// Invalid: var declaration in nested block with "both" option
		{
			Code:    "if (test) { var foo = 42; }",
			Options: []interface{}{"both"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moveDeclToRoot",
				},
			},
		},

		// Invalid: var declaration in function with nested block
		{
			Code:    "function bar() { if (test) { var foo = 42; } }",
			Options: []interface{}{"both"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moveDeclToRoot",
				},
			},
		},
	})
}
