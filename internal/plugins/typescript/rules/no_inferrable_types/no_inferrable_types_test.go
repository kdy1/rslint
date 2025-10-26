package no_inferrable_types

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoInferrableTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInferrableTypesRule,
		[]rule_tester.ValidTestCase{
			// BigInt without type annotation
			{Code: `const a = 10n;`},
			{Code: `const a = -10n;`},
			{Code: `const a = BigInt(10);`},
			{Code: `const a = -BigInt(10);`},
			{Code: `const a = BigInt?.(10);`},
			{Code: `const a = -BigInt?.(10);`},

			// Boolean without type annotation
			{Code: `const a = false;`},
			{Code: `const a = true;`},
			{Code: `const a = Boolean(null);`},
			{Code: `const a = Boolean?.(null);`},
			{Code: `const a = !0;`},

			// Number without type annotation
			{Code: `const a = 10;`},
			{Code: `const a = +10;`},
			{Code: `const a = -10;`},
			{Code: `const a = Number('1');`},
			{Code: `const a = +Number('1');`},
			{Code: `const a = -Number('1');`},
			{Code: `const a = Number?.('1');`},
			{Code: `const a = +Number?.('1');`},
			{Code: `const a = -Number?.('1');`},
			{Code: `const a = Infinity;`},
			{Code: `const a = +Infinity;`},
			{Code: `const a = -Infinity;`},
			{Code: `const a = NaN;`},
			{Code: `const a = +NaN;`},
			{Code: `const a = -NaN;`},

			// Other types
			{Code: `const a = null;`},
			{Code: `const a = /a/;`},
			{Code: `const a = RegExp('a');`},
			{Code: `const a = RegExp?.('a');`},
			{Code: `const a = new RegExp('a');`},
			{Code: `const a = 'str';`},
			{Code: "const a = `str`;"},
			{Code: `const a = String(1);`},
			{Code: `const a = String?.(1);`},
			{Code: `const a = Symbol('a');`},
			{Code: `const a = Symbol?.('a');`},
			{Code: `const a = undefined;`},
			{Code: `const a = void someValue;`},

			// Function with default params (not inferrable - complex types)
			{Code: `function fn(a = 5, b = true, c = 'foo') {}`},

			// Any type (explicit any is allowed)
			{Code: `const a: any = 5;`},
			{Code: `const fn = function (a: any = 5, b: any = true, c: any = 'foo') {};`},

			// With ignoreParameters option
			{
				Code: `const fn = (a: number = 5) => a;`,
				Options: map[string]any{"ignoreParameters": true},
			},
			{
				Code: `const fn = function (a: number = 5, b: boolean = true) {};`,
				Options: map[string]any{"ignoreParameters": true},
			},
			{
				Code: `function fn(a: number = 5, b: boolean = true) {}`,
				Options: map[string]any{"ignoreParameters": true},
			},

			// With ignoreProperties option
			{
				Code: `class Foo { a: number = 5; b: boolean = true; c: string = 'foo'; }`,
				Options: map[string]any{"ignoreProperties": true},
			},

			// Optional properties (should not be flagged)
			{Code: `class Foo { a?: number = 5; }`},
			{Code: `class Foo { b?: boolean = true; }`},

			// Constructor parameter with public modifier (needs explicit type)
			{Code: `class Foo { constructor(public a: number = 5) {} }`},
		},
		[]rule_tester.InvalidTestCase{
			// BigInt with unnecessary type annotation
			{
				Code: `const a: bigint = 10n;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = 10n;`},
			},
			{
				Code: `const a: bigint = -10n;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = -10n;`},
			},
			{
				Code: `const a: bigint = BigInt(10);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = BigInt(10);`},
			},
			{
				Code: `const a: bigint = -BigInt(10);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = -BigInt(10);`},
			},
			{
				Code: `const a: bigint = BigInt?.(10);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = BigInt?.(10);`},
			},
			{
				Code: `const a: bigint = -BigInt?.(10);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = -BigInt?.(10);`},
			},

			// Boolean with unnecessary type annotation
			{
				Code: `const a: boolean = false;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = false;`},
			},
			{
				Code: `const a: boolean = true;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = true;`},
			},

			// Number with unnecessary type annotation
			{
				Code: `const a: number = 10;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = 10;`},
			},

			// String with unnecessary type annotation
			{
				Code: `const a: string = 'str';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = 'str';`},
			},

			// Symbol with unnecessary type annotation
			{
				Code: `const a: symbol = Symbol('a');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = Symbol('a');`},
			},

			// Undefined with unnecessary type annotation
			{
				Code: `const a: undefined = undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const a = undefined;`},
			},

			// Function parameters with inferrable types
			{
				Code: `const fn = (a: number = 5, b: boolean = true, c: string = 'foo') => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
					{MessageId: "noInferrableType"},
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const fn = (a = 5, b = true, c = 'foo') => {};`},
			},
			{
				Code: `function fn(a: number = 5, b: boolean = true) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
					{MessageId: "noInferrableType"},
				},
				Output: []string{`function fn(a = 5, b = true) {}`},
			},
			{
				Code: `const fn = function (a: number = 5, b: boolean = true, c: string = 'foo') {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
					{MessageId: "noInferrableType"},
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const fn = function (a = 5, b = true, c = 'foo') {};`},
			},

			// Class properties with inferrable types
			{
				Code: `class Foo { a: number = 5; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`class Foo { a = 5; }`},
			},
			{
				Code: `class Foo { a: number = 5; b: boolean = true; c: string = 'foo'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
					{MessageId: "noInferrableType"},
					{MessageId: "noInferrableType"},
				},
				Output: []string{`class Foo { a = 5; b = true; c = 'foo'; }`},
			},

			// Optional parameter with default (should still be flagged)
			{
				Code: `const fn = (a?: number = 5) => a;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noInferrableType"},
				},
				Output: []string{`const fn = (a = 5) => a;`},
				Options: map[string]any{"ignoreParameters": false},
			},
		},
	)
}
