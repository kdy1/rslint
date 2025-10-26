package no_unreachable

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoUnreachableRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnreachableRule,
		[]rule_tester.ValidTestCase{
			// Function declarations are hoisted
			{Code: `function foo() { function bar() { return 1; } return bar(); }`},
			{Code: `function foo() { return bar(); function bar() { return 1; } }`},
			{Code: `function foo() { return x; var x; }`},

			// Variable declarations without initializers
			{Code: `function foo() { var x = 1; var y = 2; }`},
			{Code: `function foo() { var x = 1; var y = 2; return; }`},

			// Control flow
			{Code: `while (true) { switch (foo) { case 1: x = 1; x = 2;} }`},
			{Code: `while (true) { break; var x; }`},
			{Code: `while (true) { continue; var x, y; }`},
			{Code: `while (true) { throw 'message'; var x; }`},
			{Code: `while (true) { if (true) break; var x = 1; }`},
			{Code: `while (true) continue;`},

			// Switch statements
			{Code: `switch (foo) { case 1: break; var x; }`},
			{Code: `switch (foo) { case 1: break; var x; default: throw true; };`},

			// Arrow functions
			{Code: `const arrow_direction = arrow => {  switch (arrow) { default: throw new Error();  };}`},

			// Throw at top level with var
			{Code: `var x = 1; y = 2; throw 'uh oh'; var y;`},

			// Conditional returns
			{Code: `function foo() { var x = 1; if (x) { return; } x = 2; }`},
			{Code: `function foo() { var x = 1; if (x) { } else { return; } x = 2; }`},
			{Code: `function foo() { var x = 1; switch (x) { case 0: break; default: return; } x = 2; }`},

			// Loops with conditional breaks
			{Code: `function foo() { var x = 1; while (x) { return; } x = 2; }`},
			{Code: `function foo() { var x = 1; for (x in {}) { return; } x = 2; }`},
			{Code: `function foo() { var x = 1; try { return; } finally { x = 2; } }`},
			{Code: `function foo() { var x = 1; for (;;) { if (x) break; } x = 2; }`},

			// Labeled statements
			{Code: `A: { break A; } foo()`},

			// Generators and try-catch
			{Code: `function* foo() { try { yield 1; return; } catch (err) { return err; } }`},
			{Code: `function foo() { try { bar(); return; } catch (err) { return err; } }`},
			{Code: `function foo() { try { a.b.c = 1; return; } catch (err) { return err; } }`},

			// Class fields (reachable)
			{Code: `class C { foo = reachable; }`},
			{Code: `class C { foo = reachable; constructor() {} }`},
			{Code: `class C extends B { foo = reachable; }`},
			{Code: `class C extends B { foo = reachable; constructor() { super(); } }`},
			{Code: `class C extends B { static foo = reachable; constructor() {} }`},
		},
		[]rule_tester.InvalidTestCase{
			// Basic unreachable code
			{
				Code: `function foo() { return x; var x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `function foo() { return x; var x, y = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `while (true) { continue; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `function foo() { return; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `function foo() { throw error; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `while (true) { break; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `while (true) { continue; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},

			// Switch cases
			{
				Code: `function foo() { switch (foo) { case 1: return; x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `function foo() { switch (foo) { case 1: throw e; x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `while (true) { switch (foo) { case 1: break; x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `while (true) { switch (foo) { case 1: continue; x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},

			// Top level throw
			{
				Code: `var x = 1; throw 'uh oh'; var y = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},

			// Both branches terminate
			{
				Code: `function foo() { var x = 1; if (x) { return; } else { throw e; } x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `function foo() { var x = 1; if (x) return; else throw -1; x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},

			// Try-finally
			{
				Code: `function foo() { var x = 1; try { return; } finally {} x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `function foo() { var x = 1; try { } finally { return; } x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},

			// Loops
			{
				Code: `function foo() { var x = 1; do { return; } while (x); x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `function foo() { var x = 1; while (x) { if (x) break; else continue; x = 2; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `function foo() { var x = 1; for (;;) { if (x) continue; } x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `function foo() { var x = 1; while (true) { } x = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},

			// Arrow function
			{
				Code: `const arrow_direction = arrow => {  switch (arrow) { default: throw new Error();  }; g() }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 86},
				},
			},

			// Class fields (unreachable - extends without super)
			{
				Code: `class C extends B { foo; constructor() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `class C extends B { foo = unreachable + code; constructor() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
			{
				Code: `class C extends B { foo; bar; constructor() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode"},
				},
			},
		},
	)
}
