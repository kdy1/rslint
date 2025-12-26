package prefer_enum_initializers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferEnumInitializersRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferEnumInitializersRule,

		// Valid test cases - enums with all members initialized
		[]rule_tester.ValidTestCase{
			// Empty enum
			{Code: `enum Direction {}`},

			// Single member with numeric initializer
			{Code: `enum Direction { Up = 1 }`},

			// Multiple members with numeric initializers
			{Code: `enum Direction { Up = 1, Down = 2 }`},

			// Multiple members with string initializers
			{Code: `enum Direction { Up = 'Up', Down = 'Down' }`},

			// Mixed string and numeric initializers (all initialized)
			{Code: `enum Status { Open = 1, Close = 'Close' }`},

			// Computed values
			{Code: `enum E { A = 1 << 0, B = 1 << 1 }`},

			// Template literals
			{Code: "enum E { A = `a`, B = `b` }"},
		},

		// Invalid test cases - enums with uninitialized members
		[]rule_tester.InvalidTestCase{
			// Single uninitialized member
			{
				Code: `enum Direction { Up }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInitializer",
						Line:      1,
						Column:    18,
					},
				},
			},

			// Multiple uninitialized members
			{
				Code: `enum Direction { Up, Down }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInitializer",
						Line:      1,
						Column:    18,
					},
					{
						MessageId: "noInitializer",
						Line:      1,
						Column:    22,
					},
				},
			},

			// Uninitialized member after initialized member
			{
				Code: `enum Direction { Up = 'Up', Down }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInitializer",
						Line:      1,
						Column:    29,
					},
				},
			},

			// Initialized member after uninitialized member
			{
				Code: `enum Direction { Up, Down = 'Down' }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInitializer",
						Line:      1,
						Column:    18,
					},
				},
			},

			// Mix of initialized and uninitialized
			{
				Code: `enum E { A, B = 1, C }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInitializer",
						Line:      1,
						Column:    10,
					},
					{
						MessageId: "noInitializer",
						Line:      1,
						Column:    20,
					},
				},
			},

			// Multi-line enum with uninitialized members
			{
				Code: `
enum Status {
  Open,
  Close = 'Close',
  Pending
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInitializer",
						Line:      3,
						Column:    3,
					},
					{
						MessageId: "noInitializer",
						Line:      5,
						Column:    3,
					},
				},
			},

			// String literal member names
			{
				Code: `enum E { "a" }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInitializer",
						Line:      1,
						Column:    10,
					},
				},
			},

			// Numeric literal member names
			{
				Code: `enum E { 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noInitializer",
						Line:      1,
						Column:    10,
					},
				},
			},
		},
	)
}
