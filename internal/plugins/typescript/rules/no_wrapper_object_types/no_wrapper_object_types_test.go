package no_wrapper_object_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoWrapperObjectTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoWrapperObjectTypesRule, []rule_tester.ValidTestCase{
		// Valid primitive types
		{Code: `let value: bigint;`},
		{Code: `let value: boolean;`},
		{Code: `let value: never;`},
		{Code: `let value: null;`},
		{Code: `let value: number;`},
		{Code: `let value: symbol;`},
		{Code: `let value: undefined;`},
		{Code: `let value: unknown;`},
		{Code: `let value: void;`},

		// Valid function types
		{Code: `let value: () => void;`},
		{Code: `let value: () => () => void;`},

		// Valid: Custom identifiers matching wrapper names
		{Code: `let Bigint = 3;`},
		{Code: `interface Boolean {}`},
		{Code: `type Number = {};`},
		{Code: `type Number = 0 | 1;`},

		// Valid: Generic type parameters
		{Code: `function foo<T = Boolean>() {}`},
		{Code: `function foo<T = Number>() {}`},
		{Code: `function foo<T = String>() {}`},

		// Valid: Class extending Number (unusual but syntactically valid)
		{Code: `class Foo extends Number {}`},

		// Valid: lowercase object type
		{Code: `let value: object;`},
	}, []rule_tester.InvalidTestCase{
		// Invalid: BigInt wrapper
		{
			Code:   `let value: BigInt;`,
			Output: []string{`let value: bigint;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    12,
				},
			},
		},
		// Invalid: Boolean wrapper
		{
			Code:   `let value: Boolean;`,
			Output: []string{`let value: boolean;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    12,
				},
			},
		},
		// Invalid: Number wrapper
		{
			Code:   `let value: Number;`,
			Output: []string{`let value: number;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    12,
				},
			},
		},
		// Invalid: Object wrapper
		{
			Code:   `let value: Object;`,
			Output: []string{`let value: object;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    12,
				},
			},
		},
		// Invalid: String wrapper
		{
			Code:   `let value: String;`,
			Output: []string{`let value: string;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    12,
				},
			},
		},
		// Invalid: Symbol wrapper
		{
			Code:   `let value: Symbol;`,
			Output: []string{`let value: symbol;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    12,
				},
			},
		},
		// Invalid: Multiple wrapper types in union
		{
			Code:   `let value: Number | Symbol;`,
			Output: []string{`let value: number | Symbol;`, `let value: Number | symbol;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    12,
				},
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    21,
				},
			},
		},
		// Invalid: Type assertion
		{
			Code:   `0 as Number;`,
			Output: []string{`0 as number;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    6,
				},
			},
		},
		// Invalid: Type alias
		{
			Code:   `type MyType = Number;`,
			Output: []string{`type MyType = number;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    15,
				},
			},
		},
		// Invalid: Tuple element type
		{
			Code:   `type MyType = [Number];`,
			Output: []string{`type MyType = [number];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    16,
				},
			},
		},
		// Invalid: Intersection types
		{
			Code:   `type MyType = Number & String;`,
			Output: []string{`type MyType = number & String;`, `type MyType = Number & string;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    15,
				},
				{
					MessageId: "bannedClassType",
					Line:      1,
					Column:    24,
				},
			},
		},
	})
}
