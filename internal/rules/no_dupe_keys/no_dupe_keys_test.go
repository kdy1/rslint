package no_dupe_keys

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoDupeKeysRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeKeysRule,
		[]rule_tester.ValidTestCase{
			// Basic valid cases
			{Code: `var foo = { __proto__: 1, two: 2};`},
			{Code: `var x = { foo: 1, bar: 2 };`},
			{Code: `var x = { '': 1, bar: 2 };`},
			{Code: `var x = { '': 1, ' ': 2 };`},
			{Code: `var x = { foo: 1, bar: 2 };`},

			// Getter and setter for same property
			{Code: `var x = { get a() {}, set a(b) {} };`},
			{Code: `var x = { a: 1, get b() {}, set b(c) {} };`},

			// Computed properties (ES6)
			{Code: `var foo = { ['bar']: 1, baz: 2 };`},
			{Code: `var x = { a: b, [a]: b };`},
			{Code: `var x = { [a]: b, [a]: c };`}, // Computed names can't be statically checked

			// __proto__ special cases
			{Code: `var x = { '__proto__': 1, bar: 2 };`},
			{Code: `var x = { ['__proto__']: 1, bar: 2 };`},

			// Numeric keys
			{Code: `var x = { 0o12: 1, 12: 2 };`}, // Octal 0o12 is decimal 10, different from 12
			{Code: `var x = { 1_0: 1, 1_000: 2 };`}, // Different numbers with separators
			{Code: `var x = { 1_0: 1, 1_000: 2 };`},
		},
		[]rule_tester.InvalidTestCase{
			// String key duplicates identifier key
			{
				Code: `var foo = { bar: 1, 'bar': 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},

			// Duplicate property y
			{
				Code: `var x = { a: 1, b: 2, a: 3 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},

			// Duplicate empty string keys
			{
				Code: `var x = { "": 0, "": 1 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},

			// Hexadecimal 0x1 vs decimal 1
			{
				Code: `var x = { 0x1: 1, 1: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},

			// Octal 0o12 vs decimal 10
			{
				Code: `var x = { 0o12: 1, 10: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},

			// Binary 0b1 vs decimal 1
			{
				Code: `var x = { 0b1: 1, 1: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},

			// Octal 0o1 vs decimal 1
			{
				Code: `var x = { 0o1: 1, 1: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},

			// BigInt 1n vs number 1
			{
				Code: `var x = { 1n: 1, 1: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},

			// Numeric separator 1_0 vs 10
			{
				Code: `var x = { 1_0: 1, 10: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 19},
				},
			},

			// String "z" vs identifier z
			{
				Code: `var foo = { "z": 1, z: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},

			// Property a with duplicate getter
			{
				Code: `var x = { a: 1, get a() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},

			// Property a with duplicate setter
			{
				Code: `var x = { a: 1, set a(b) {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},

			// Duplicate __proto__ literal keys
			{
				Code: `var x = { __proto__: 1, __proto__: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},

			// Duplicate getter
			{
				Code: `var x = { get a() {}, get a() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 27},
				},
			},

			// Duplicate setter
			{
				Code: `var x = { set a(b) {}, set a(c) {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
		},
	)
}
