package unified_signatures

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestUnifiedSignaturesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&UnifiedSignaturesRule,
		[]rule_tester.ValidTestCase{
			// No error for arity difference greater than 1
			{Code: `
interface I {
  a2(): void;
  a2(x: number, y: number): void;
}
			`},

			// No error for different return types
			{Code: `
interface I {
  a4(): void;
  a4(x: number): number;
}
			`},

			// No error if one takes a type parameter and the other doesn't
			{Code: `
interface I {
  a5<T>(x: T): T;
  a5(x: number): number;
}
			`},

			// No error if one is a rest parameter and other isn't
			{Code: `
interface I {
  b2(x: string): void;
  b2(...x: number[]): void;
}
			`},

			// No error if both are rest parameters
			{Code: `
interface I {
  b3(...x: number[]): void;
  b3(...x: string[]): void;
}
			`},

			// No error if one is optional and the other isn't
			{Code: `
interface I {
  c3(x: number): void;
  c3(x?: string): void;
}
			`},

			// No error if they differ by 2 or more parameters
			{Code: `
interface I {
  d2(x: string, y: number): void;
  d2(x: number, y: string): void;
}
			`},

			// No conflict between static/non-static members
			{Code: `
declare class D {
  static a();
  a(x: number);
}
			`},

			// Allow separate overloads if one is generic and the other isn't
			{Code: `
interface Generic<T> {
  x(): void;
  x(x: T[]): void;
}
			`},

			// Allow signatures if the type is not equal
			{Code: `
interface I {
  f(x1: number): void;
  f(x1: boolean, x2?: number): void;
}
			`},

			// Allow type parameters that are not equal
			{Code: `
function f<T extends number>(x: T[]): void;
function f<T extends string>(x: T): void;
			`},

			// With ignoreDifferentlyNamedParameters option
			{
				Code: `
function f(a: number): void;
function f(b: string): void;
function f(a: number | string): void {}
				`,
				Options: map[string]interface{}{
					"ignoreDifferentlyNamedParameters": true,
				},
			},

			// With ignoreOverloadsWithDifferentJSDoc option
			{
				Code: `
/** @deprecated */
declare function f(x: number): unknown;
declare function f(x: boolean): unknown;
				`,
				Options: map[string]interface{}{
					"ignoreOverloadsWithDifferentJSDoc": true,
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Single parameter type difference
			{
				Code: `
function f(a: number): void;
function f(b: string): void;
function f(a: number | string): void {}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "singleParameterDifference",
						Line:      3,
						Column:    12,
					},
				},
			},

			// Omitting single parameter
			{
				Code: `
function opt(xs?: number[]): void;
function opt(xs: number[], y: string): void;
function opt(...args: any[]) {}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "omittingSingleParameter",
						Line:      3,
						Column:    28,
					},
				},
			},

			// Error for extra parameter
			{
				Code: `
interface I {
  a1(): void;
  a1(x: number): void;
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "omittingSingleParameter",
						Line:      4,
						Column:    6,
					},
				},
			},

			// Error if only one defines a rest parameter
			{
				Code: `
interface I {
  b(): void;
  b(...x: number[]): void;
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "omittingRestParameter",
						Line:      4,
						Column:    5,
					},
				},
			},

			// Error if only one defines an optional parameter
			{
				Code: `
interface I {
  c(): void;
  c(x?: number): void;
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "omittingSingleParameter",
						Line:      4,
						Column:    5,
					},
				},
			},

			// Error if both are optional
			{
				Code: `
interface I {
  c2(x?: number): void;
  c2(x?: string): void;
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "singleParameterDifference",
						Line:      4,
						Column:    6,
					},
				},
			},

			// Error for different types (could be a union)
			{
				Code: `
interface I {
  d(x: number): void;
  d(x: string): void;
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "singleParameterDifference",
						Line:      4,
						Column:    5,
					},
				},
			},

			// Works for type literal and call signature too
			{
				Code: `
type T = {
  (): void;
  (x: number): void;
};
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "omittingSingleParameter",
						Line:      4,
						Column:    4,
					},
				},
			},

			// Works for constructor
			{
				Code: `
declare class C {
  constructor();
  constructor(x: number);
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "omittingSingleParameter",
						Line:      4,
						Column:    15,
					},
				},
			},

			// Works with unions
			{
				Code: `
interface I {
  f(x: number);
  f(x: string | boolean);
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "singleParameterDifference",
						Line:      4,
						Column:    5,
					},
				},
			},

			// Check type parameters when equal
			{
				Code: `
function f<T>(x: T[]): void;
function f<T>(x: T): void;
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "singleParameterDifference",
						Line:      3,
						Column:    15,
					},
				},
			},

			// Verifies type parameters and constraints
			{
				Code: `
function f<T extends number>(x: T[]): void;
function f<T extends number>(x: T): void;
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "singleParameterDifference",
						Line:      3,
						Column:    30,
					},
				},
			},

			// Works with abstract
			{
				Code: `
abstract class Foo {
  public abstract f(x: number): void;
  public abstract f(x: string): void;
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "singleParameterDifference",
						Line:      4,
						Column:    21,
					},
				},
			},

			// Works with new constructor
			{
				Code: `
interface Foo {
  new (x: string): Foo;
  new (x: number): Foo;
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "singleParameterDifference",
						Line:      4,
						Column:    8,
					},
				},
			},

			// Export function overloads
			{
				Code: `
export function foo(line: number): number;
export function foo(line: number, character?: number): number;
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "omittingSingleParameter",
						Line:      3,
						Column:    35,
					},
				},
			},

			// This parameter type difference
			{
				Code: `
function f(this: string): void;
function f(this: number): void;
function f(this: string | number): void {}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "singleParameterDifference",
						Line:      3,
						Column:    12,
					},
				},
			},

			// This parameter with regular parameters
			{
				Code: `
function f(this: string, a: boolean): void;
function f(this: number, a: boolean): void;
function f(this: string | number, a: boolean): void {}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "singleParameterDifference",
						Line:      3,
						Column:    12,
					},
				},
			},
		},
	)
}
