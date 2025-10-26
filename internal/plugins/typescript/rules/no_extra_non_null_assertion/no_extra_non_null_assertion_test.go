package no_extra_non_null_assertion

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoExtraNonNullAssertionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExtraNonNullAssertionRule,
		[]rule_tester.ValidTestCase{
			// Single non-null assertion is valid
			{Code: `
const foo = { bar: null as number | null };
foo!.bar;
`},
			{Code: `
function foo(bar: number | null | undefined) {
  return bar!;
}
`},
			{Code: `
const foo = { bar: null as string | null };
foo.bar!.trim();
`},
			// Optional chaining without non-null assertion is valid
			{Code: `
const foo = { bar: null as { baz?: number } | null };
foo?.bar?.baz;
`},
			// GitHub issue #2166 - optional chaining with assertion on the right side
			{Code: `
const checksCounter: { textContent: string | null } | null = null;
checksCounter?.textContent!.trim();
`},
			// GitHub issue #2732 - non-null assertion within optional bracket notation
			{Code: `
const foo: { [key: string]: string | null } | null = null;
foo?.['bar']!;
`},
		},
		[]rule_tester.InvalidTestCase{
			// Double non-null assertion
			{
				Code: `
const foo = { bar: null as number | null };
foo!!.bar;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noExtraNonNullAssertion",
						Line:      3,
						Column:    1,
					},
				},
				Output: []string{`
const foo = { bar: null as number | null };
foo!.bar;
`},
			},
			{
				Code: `
function foo(bar: number | null | undefined) {
  return bar!!;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noExtraNonNullAssertion",
						Line:      3,
						Column:    10,
					},
				},
				Output: []string{`
function foo(bar: number | null | undefined) {
  return bar!;
}
`},
			},
			// Non-null assertion before optional chaining operator
			{
				Code: `
const foo = { bar: null as { baz?: number } | null };
foo!?.bar;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noExtraNonNullAssertion",
						Line:      3,
						Column:    1,
					},
				},
				Output: []string{`
const foo = { bar: null as { baz?: number } | null };
foo?.bar;
`},
			},
			// Non-null assertion before optional call operator
			{
				Code: `
const foo = null as (() => void) | null;
foo!?.();
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noExtraNonNullAssertion",
						Line:      3,
						Column:    1,
					},
				},
				Output: []string{`
const foo = null as (() => void) | null;
foo?.();
`},
			},
			// Double assertion with parentheses
			{
				Code: `
const foo = { bar: null as number | null };
(foo)!!.bar;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noExtraNonNullAssertion",
						Line:      3,
						Column:    1,
					},
				},
				Output: []string{`
const foo = { bar: null as number | null };
(foo)!.bar;
`},
			},
			// Parenthesized non-null assertion before optional chaining
			{
				Code: `
const foo = { bar: null as { baz?: number } | null };
(foo)!?.bar;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noExtraNonNullAssertion",
						Line:      3,
						Column:    1,
					},
				},
				Output: []string{`
const foo = { bar: null as { baz?: number } | null };
(foo)?.bar;
`},
			},
			// Parenthesized expression with optional call
			{
				Code: `
const foo = null as (() => void) | null;
(foo)!?.();
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noExtraNonNullAssertion",
						Line:      3,
						Column:    1,
					},
				},
				Output: []string{`
const foo = null as (() => void) | null;
(foo)?.();
`},
			},
			// Multiple exclamation marks
			{
				Code: `
const foo = null as number | null;
foo!!!;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noExtraNonNullAssertion",
						Line:      3,
						Column:    1,
					},
					{
						MessageId: "noExtraNonNullAssertion",
						Line:      3,
						Column:    1,
					},
				},
				Output: []string{`
const foo = null as number | null;
foo!;
`},
			},
			// Non-null before optional element access
			{
				Code: `
const foo: { [key: string]: number } | null = null;
foo!?.[0];
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noExtraNonNullAssertion",
						Line:      3,
						Column:    1,
					},
				},
				Output: []string{`
const foo: { [key: string]: number } | null = null;
foo?.[0];
`},
			},
		},
	)
}
