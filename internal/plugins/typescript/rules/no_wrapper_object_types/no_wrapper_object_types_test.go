package no_wrapper_object_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoWrapperObjectTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoWrapperObjectTypesRule, []rule_tester.ValidTestCase{
		{Code: "let value: NumberLike;"},
		{Code: "let value: Other;"},
		{Code: "let value: bigint;"},
		{Code: "let value: boolean;"},
		{Code: "let value: never;"},
		{Code: "let value: null;"},
		{Code: "let value: number;"},
		{Code: "let value: symbol;"},
		{Code: "let value: undefined;"},
		{Code: "let value: unknown;"},
		{Code: "let value: void;"},
		{Code: "let value: () => void;"},
		{Code: "let value: () => () => void;"},
		{Code: "let Bigint;"},
		{Code: "let Boolean;"},
		{Code: "let Never;"},
		{Code: "let Null;"},
		{Code: "let Number;"},
		{Code: "let Symbol;"},
		{Code: "let Undefined;"},
		{Code: "let Unknown;"},
		{Code: "let Void;"},
		{Code: "interface Bigint {}"},
		{Code: "interface Boolean {}"},
		{Code: "interface Never {}"},
		{Code: "interface Null {}"},
		{Code: "interface Number {}"},
		{Code: "interface Symbol {}"},
		{Code: "interface Undefined {}"},
		{Code: "interface Unknown {}"},
		{Code: "interface Void {}"},
		{Code: "type Bigint = {};"},
		{Code: "type Boolean = {};"},
		{Code: "type Never = {};"},
		{Code: "type Null = {};"},
		{Code: "type Number = {};"},
		{Code: "type Symbol = {};"},
		{Code: "type Undefined = {};"},
		{Code: "type Unknown = {};"},
		{Code: "type Void = {};"},
		{Code: "class MyClass extends Number {}"},
		{Code: `
type Number = 0 | 1;
let value: Number;
`},
		{Code: `
type Bigint = 0 | 1;
let value: Bigint;
`},
		{Code: `
type T<Symbol> = Symbol;
type U<UU> = UU extends T<infer Function> ? Function : never;
`},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "let value: BigInt;",
			Output: []string{"let value: bigint;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    12,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
		{
			Code:   "let value: Boolean;",
			Output: []string{"let value: boolean;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    12,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
		{
			Code:   "let value: Number;",
			Output: []string{"let value: number;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    12,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
		{
			Code:   "let value: Object;",
			Output: []string{"let value: object;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    12,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
		{
			Code:   "let value: String;",
			Output: []string{"let value: string;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    12,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
		{
			Code:   "let value: Symbol;",
			Output: []string{"let value: symbol;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    12,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
		{
			Code:   "let value: Number | Symbol;",
			Output: []string{"let value: number | symbol;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    12,
					Line:      1,
					MessageId: "bannedClassType",
				},
				{
					Column:    21,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
		{
			Code:   "let value: { property: Number };",
			Output: []string{"let value: { property: number };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    24,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
		{
			Code:   "0 as Number;",
			Output: []string{"0 as number;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    6,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
		{
			Code:   "type MyType = Number;",
			Output: []string{"type MyType = number;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    15,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
		{
			Code:   "type MyType = [Number];",
			Output: []string{"type MyType = [number];"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    16,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
		{
			Code: "class MyClass implements Number {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{Column: 26, Line: 1, MessageId: "bannedClassType"},
			},
		},
		{
			Code: "interface MyInterface extends Number {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{Column: 31, Line: 1, MessageId: "bannedClassType"},
			},
		},
		{
			Code:   "type MyType = Number & String;",
			Output: []string{"type MyType = number & string;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Column:    15,
					Line:      1,
					MessageId: "bannedClassType",
				},
				{
					Column:    24,
					Line:      1,
					MessageId: "bannedClassType",
				},
			},
		},
	})
}
