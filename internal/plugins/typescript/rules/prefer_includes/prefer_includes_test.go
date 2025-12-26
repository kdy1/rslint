package prefer_includes

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferIncludesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferIncludesRule, []rule_tester.ValidTestCase{
		// indexOf without comparison
		{Code: "foo.indexOf(bar)"},
		{Code: "foo.indexOf(bar) + 0"},

		// Union types
		{Code: `
declare const val: string | number;
val.indexOf('foo') === -1;
		`},

		// User-defined types without includes
		{Code: `
type UserType = { indexOf(x: string): number };
declare const foo: UserType;
foo.indexOf('x') === -1;
		`},

		// User-defined types with different includes signature
		{Code: `
type WithIncludesButDifferent = {
	indexOf(x: string): number;
	includes(x: string, fromIndex: number): boolean;
};
declare const foo: WithIncludesButDifferent;
foo.indexOf('x') === -1;
		`},

		// includes is a boolean, not a method
		{Code: `
type WithBooleanIncludes = {
	indexOf(x: string): number;
	includes: boolean;
};
declare const foo: WithBooleanIncludes;
foo.indexOf('x') === -1;
		`},

		// RegExp with flags
		{Code: `/bar/i.test(foo)`},

		// RegExp with character class
		{Code: `/ba[rz]/.test(foo)`},

		// RegExp with alternation
		{Code: `/foo|bar/.test(foo)`},

		// RegExp test without arguments
		{Code: `pattern.test()`},

		// Non-RegExp test method
		{Code: `
type WithTest = { test(x: string): boolean };
declare const obj: WithTest;
obj.test(foo);
		`},

		// RegExp from variable
		{Code: `
const pattern = /foo/;
const regex = pattern;
regex.test(str);
		`},
	}, []rule_tester.InvalidTestCase{
		// Basic indexOf comparisons - positive checks (should use includes)
		{
			Code: `foo.indexOf(bar) !== -1`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`foo.includes(bar)`},
		},
		{
			Code: `foo.indexOf(bar) != -1`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`foo.includes(bar)`},
		},
		{
			Code: `foo.indexOf(bar) > -1`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`foo.includes(bar)`},
		},
		{
			Code: `foo.indexOf(bar) >= 0`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`foo.includes(bar)`},
		},

		// Basic indexOf comparisons - negative checks (should use !includes)
		{
			Code: `foo.indexOf(bar) === -1`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`!foo.includes(bar)`},
		},
		{
			Code: `foo.indexOf(bar) == -1`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`!foo.includes(bar)`},
		},
		{
			Code: `foo.indexOf(bar) < 0`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`!foo.includes(bar)`},
		},
		{
			Code: `foo.indexOf(bar) <= -1`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`!foo.includes(bar)`},
		},

		// Reversed operands
		{
			Code: `-1 !== foo.indexOf(bar)`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`foo.includes(bar)`},
		},
		{
			Code: `0 <= foo.indexOf(bar)`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`foo.includes(bar)`},
		},
		{
			Code: `-1 === foo.indexOf(bar)`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`!foo.includes(bar)`},
		},
		{
			Code: `0 > foo.indexOf(bar)`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
			Output: []string{`!foo.includes(bar)`},
		},

		// Optional chaining - should report but not fix
		{
			Code: `a?.indexOf(b) === -1`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
		},
		{
			Code: `a?.indexOf(b) !== -1`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 1, Column: 1},
			},
		},

		// RegExp.test() cases
		{
			Code: `/bar/.test(foo)`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStringIncludes", Line: 1, Column: 1},
			},
			Output: []string{`foo.includes("bar")`},
		},
		{
			Code: `/bar/.test((1 + 1, foo))`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStringIncludes", Line: 1, Column: 1},
			},
			Output: []string{`(1 + 1, foo).includes("bar")`},
		},
		{
			Code: `/\\0'\\\\\\n\\r\\v\\t\\f/.test(foo)`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStringIncludes", Line: 1, Column: 1},
			},
			Output: []string{`foo.includes("\\0'\\\\\\n\\r\\v\\t\\f")`},
		},
		{
			Code: `new RegExp('bar').test(foo)`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStringIncludes", Line: 1, Column: 1},
			},
			Output: []string{`foo.includes("bar")`},
		},
		{
			Code: `
const pattern = 'bar';
new RegExp(pattern).test(foo + bar);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferStringIncludes", Line: 3, Column: 1},
			},
			Output: []string{`
const pattern = 'bar';
(foo + bar).includes("bar");
			`},
		},

		// Array types
		{
			Code: `
declare const arr: any[];
arr.indexOf(x) !== -1;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 3, Column: 1},
			},
			Output: []string{`
declare const arr: any[];
arr.includes(x);
			`},
		},
		{
			Code: `
declare const arr: ReadonlyArray<any>;
arr.indexOf(x) !== -1;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 3, Column: 1},
			},
			Output: []string{`
declare const arr: ReadonlyArray<any>;
arr.includes(x);
			`},
		},

		// TypedArrays
		{
			Code: `
declare const arr: Int8Array;
arr.indexOf(x) !== -1;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 3, Column: 1},
			},
			Output: []string{`
declare const arr: Int8Array;
arr.includes(x);
			`},
		},
		{
			Code: `
declare const arr: Uint8Array;
arr.indexOf(x) !== -1;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 3, Column: 1},
			},
			Output: []string{`
declare const arr: Uint8Array;
arr.includes(x);
			`},
		},
		{
			Code: `
declare const arr: Float32Array;
arr.indexOf(x) !== -1;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 3, Column: 1},
			},
			Output: []string{`
declare const arr: Float32Array;
arr.includes(x);
			`},
		},

		// Generic constraints
		{
			Code: `
function fn<T extends string>(x: T) {
	return x.indexOf('a') !== -1;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 3, Column: 9},
			},
			Output: []string{`
function fn<T extends string>(x: T) {
	return x.includes('a');
}
			`},
		},

		// Readonly arrays
		{
			Code: `
declare const arr: Readonly<any[]>;
arr.indexOf(x) !== -1;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 3, Column: 1},
			},
			Output: []string{`
declare const arr: Readonly<any[]>;
arr.includes(x);
			`},
		},

		// User-defined types with matching indexOf and includes
		{
			Code: `
type WithBothMethods = {
	indexOf(x: string): number;
	includes(x: string): boolean;
};
declare const foo: WithBothMethods;
foo.indexOf('x') !== -1;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIncludes", Line: 7, Column: 1},
			},
			Output: []string{`
type WithBothMethods = {
	indexOf(x: string): number;
	includes(x: string): boolean;
};
declare const foo: WithBothMethods;
foo.includes('x');
			`},
		},
	})
}
