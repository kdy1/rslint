package no_unnecessary_type_parameters

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTypeParametersRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryTypeParametersRule,

		// Valid test cases - type parameters used multiple times
		[]rule_tester.ValidTestCase{
			// Identity function - T used in parameter and return type
			{Code: `function identity<T>(arg: T): T { return arg; }`},

			// Type parameter used in property access
			{Code: `function printProperty<T>(obj: T, key: keyof T) { console.log(obj[key]); }`},

			// Type parameter with constraint used multiple times
			{Code: `function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] { return obj[key]; }`},

			// Class with type parameter used in properties
			{Code: `class Box<T> { value: T; constructor(v: T) { this.value = v; } }`},

			// Arrow function with multiple uses
			{Code: `const func = <T,>(a: T, b: T) => a;`},

			// Method declaration
			{Code: `class C { method<T>(a: T): T { return a; } }`},

			// Type parameter used in return type wrapper with explicit type
			{Code: `function box<T>(val: T): { val: T } { return { val }; }`},

			// Multiple type parameters both used
			{Code: `function map<K, V>(m: Map<K, V>, key: K): V | undefined { return m.get(key); }`},

			// Type guard
			{Code: `function isNonNull<T>(v: T): v is Exclude<T, null> { return v !== null; }`},

			// Constrained type parameter used multiple times
			{Code: `function lengthyIdentity<T extends { length: number }>(x: T): T { return x; }`},

			// Array type used twice
			{Code: `function arrayOfPairs<T>(): [T, T][] { return []; }`},

			// Generic constraint
			{Code: `function doStuff<T extends string>(x: T): T { return x; }`},
		},

		// Invalid test cases - type parameters used once or not at all
		[]rule_tester.InvalidTestCase{
			// Type parameter never used
			{
				Code: `function test<T>() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sole"},
				},
			},

			// Type parameter used only in parameter
			{
				Code: `const func = <T,>(param: T) => null;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sole"},
				},
			},

			// Type parameter used only in array parameter
			{
				Code: `const func = <T,>(param: T[]) => null;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sole"},
				},
			},

			// Type parameter used only in return type
			{
				Code: `function get<T>(): T { return null as any; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sole"},
				},
			},

			// Constrained type parameter used only once
			{
				Code: `function get<T extends object>(): T { return {} as T; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sole"},
				},
			},

			// Class with unused type parameter
			{
				Code: `class C<T> {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sole"},
				},
			},

			// Method with type parameter used once
			{
				Code: `class C { method<T>(x: T) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sole"},
				},
			},

			// Arrow function with constraint
			{
				Code: `const func = <T extends string | number>(x: T[]) => null;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sole"},
				},
			},

			// Multiple type parameters, one unused
			{
				Code: `function test<A, B>(a: A, b: A): A { return a; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sole"},
				},
			},

			// Function expression
			{
				Code: `const fn = function<T>(param: T) {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sole"},
				},
			},
		},
	)
}
