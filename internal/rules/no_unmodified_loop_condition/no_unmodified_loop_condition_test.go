package no_unmodified_loop_condition

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoUnmodifiedLoopConditionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnmodifiedLoopConditionRule,
		[]rule_tester.ValidTestCase{
			// Valid cases - loop condition is modified
			{Code: "var foo = 0; while (foo) { ++foo; }"},
			{Code: "let foo = 0; while (foo) { ++foo; }"},
			{Code: "var foo = 0; while (foo) { foo += 1; }"},
			{Code: "var foo = 0; while (foo++) { }"},
			{Code: "var foo = 0; while (foo = next()) { }"},
			{Code: "var foo = 0; while (ok(foo)) { }"},
			{Code: "var foo = 0, bar = 0; while (++foo < bar) { }"},

			// Property access - dynamic
			{Code: "var foo = 0; while (foo.ok) { }"},

			// Function call in condition
			{Code: "var foo = 0; while (foo) { update(); } function update() { ++foo; }"},

			// Multiple variables
			{Code: "var foo = 0, bar = 9; while (foo < bar) { foo += 1; }"},
			{Code: "var foo = 0, bar = 1, baz = 2; while (foo ? bar : baz) { foo += 1; }"},
			{Code: "var foo = 0, bar = 0; while (foo && bar) { ++foo; ++bar; }"},
			{Code: "var foo = 0, bar = 0; while (foo || bar) { ++foo; ++bar; }"},

			// do-while loops
			{Code: "var foo = 0; do { ++foo; } while (foo);"},
			{Code: "var foo = 0; do { } while (foo++);"},

			// for loops
			{Code: "for (var foo = 0; foo; ++foo) { }"},
			{Code: "for (var foo = 0; foo;) { ++foo }"},
			{Code: "var foo = 0, bar = 0; for (bar; foo;) { ++foo }"},

			// Not a loop
			{Code: "var foo; if (foo) { }"},

			// Array length pattern
			{Code: "var a = [1, 2, 3]; var len = a.length; for (var i = 0; i < len - 1; i++) {}"},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid cases - loop condition is not modified
			{
				Code: "var foo = 0; while (foo) { } foo = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified"},
				},
			},
			{
				Code: "var foo = 0; while (!foo) { } foo = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified"},
				},
			},
			{
				Code: "var foo = 0; while (foo != null) { } foo = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified"},
				},
			},
			{
				Code: "var foo = 0, bar = 9; while (foo < bar) { } foo = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified"},
				},
			},
			{
				Code: "var foo = 0, bar = 0; while (foo && bar) { ++bar; } foo = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified"},
				},
			},
			{
				Code: "var foo = 0, bar = 0; while (foo && bar) { ++foo; } foo = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified"},
				},
			},
			{
				Code: "var foo = 0; while (foo ? 1 : 0) { } foo = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified"},
				},
			},

			// Function with same parameter name doesn't modify outer var
			{
				Code: "var foo = 0; while (foo) { update(); } function update(foo) { ++foo; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified"},
				},
			},

			// do-while
			{
				Code: "var foo; do { } while (foo);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified"},
				},
			},

			// for loop
			{
				Code: "for (var foo = 0; foo < 10; ) { } foo = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "loopConditionNotModified"},
				},
			},
		},
	)
}
