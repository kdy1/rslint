package no_dupe_keys

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeKeysRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeKeysRule,
		[]rule_tester.ValidTestCase{
			// Valid cases - no duplicate keys
			{Code: `var foo = { __proto__: 1, two: 2};`},
			{Code: `var x = { foo: 1, bar: 2 };`},
			{Code: `var x = { '': 1, bar: 2 };`},
			{Code: `var x = { '': 1, ' ': 2 };`},
			{Code: `var x = { '': 1, [null]: 2 };`},
			{Code: `var x = { '': 1, [a]: 2 };`},
			{Code: `var x = { [a]: 1, [a]: 2 };`}, // Dynamic keys can't be checked
			{Code: `+{ get a() { }, set a(b) { } };`},
			{Code: `var x = { a: b, [a]: b };`}, // One literal, one computed
			{Code: `var x = { a: b, ...c }`},
			{Code: `var x = { get a() {}, set a (value) {} };`},
			{Code: `var x = { a: 1, b: { a: 2 } };`},
			{Code: `var {a, a} = obj`}, // Destructuring, not object literal
			{Code: `var x = { __proto__: null, ['__proto__']: null };`},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid cases - duplicate keys
			{
				Code: `var x = { a: b, ['a']: b };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "a",
						},
					},
				},
			},
			{
				Code: `var x = { y: 1, y: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "y",
						},
					},
				},
			},
			{
				Code: `var x = { '': 1, '': 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "",
						},
					},
				},
			},
			{
				Code: "var x = { '': 1, [``]: 2 };",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "",
						},
					},
				},
			},
			{
				Code: `var foo = { 0x1: 1, 1: 2};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "1",
						},
					},
				},
			},
			{
				Code: `var x = { 012: 1, 10: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "10",
						},
					},
				},
			},
			{
				Code: `var x = { 0b1: 1, 1: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "1",
						},
					},
				},
			},
			{
				Code: `var x = { 0o1: 1, 1: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "1",
						},
					},
				},
			},
			{
				Code: `var x = { "z": 1, z: 2 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "z",
						},
					},
				},
			},
			{
				Code: `var foo = { bar: 1, bar: 1, }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "bar",
						},
					},
				},
			},
			{
				Code: `var x = { a: 1, get a() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "a",
						},
					},
				},
			},
			{
				Code: `var x = { a: 1, set a(value) {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "a",
						},
					},
				},
			},
		},
	)
}
