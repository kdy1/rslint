package prefer_string_starts_ends_with

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStringStartsEndsWithRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferStringStartsEndsWithRule,
		[]rule_tester.ValidTestCase{
			// String array indexing (not string indexing)
			{Code: `function f(s: string[]) { s[0] === 'a'; }`},
			{Code: `function f(s: string[] | null) { s?.[0] === 'a'; }`},
			{Code: `function f(s: string[] | undefined) { s?.[0] === 'a'; }`},

			// Not equality comparison
			{Code: `function f(s: string) { s[0] + 'a'; }`},

			// Not index 0
			{Code: `function f(s: string) { s[1] === 'a'; }`},
			{Code: `function f(s: string | undefined) { s?.[1] === 'a'; }`},

			// Union with array
			{Code: `function f(s: string | string[]) { s[0] === 'a'; }`},

			// any type
			{Code: `function f(s: any) { s[0] === 'a'; }`},

			// Generic type
			{Code: `function f<T>(s: T) { s[0] === 'a'; }`},

			// Array indexing with length - 1
			{Code: `function f(s: string[]) { s[s.length - 1] === 'a'; }`},
			{Code: `function f(s: string[] | undefined) { s?.[s.length - 1] === 'a'; }`},

			// Not length - 1
			{Code: `function f(s: string) { s[s.length - 2] === 'a'; }`},
			{Code: `function f(s: string | undefined) { s?.[s.length - 2] === 'a'; }`},

			// charAt on arrays
			{Code: `function f(s: string[]) { s.charAt(0) === 'a'; }`},
			{Code: `function f(s: string[] | undefined) { s?.charAt(0) === 'a'; }`},

			// charAt not used for comparison
			{Code: `function f(s: string) { s.charAt(0) + 'a'; }`},

			// charAt with index other than 0
			{Code: `function f(s: string) { s.charAt(1) === 'a'; }`},
			{Code: `function f(s: string | undefined) { s?.charAt(1) === 'a'; }`},

			// charAt without argument
			{Code: `function f(s: string) { s.charAt() === 'a'; }`},

			// charAt on arrays with length - 1
			{Code: `function f(s: string[]) { s.charAt(s.length - 1) === 'a'; }`},

			// charAt with different object lengths
			{Code: `function f(a: string, b: string, c: string) { (a + b).charAt((a + c).length - 1) === 'a'; }`},
			{Code: `function f(a: string, b: string, c: string) { (a + b).charAt(c.length - 1) === 'a'; }`},

			// indexOf on arrays
			{Code: `function f(s: string[]) { s.indexOf(needle) === 0; }`},
			{Code: `function f(s: string | string[]) { s.indexOf(needle) === 0; }`},

			// indexOf comparing to wrong value
			{Code: `function f(s: string) { s.indexOf(needle) === s.length - needle.length; }`},

			// lastIndexOf on arrays
			{Code: `function f(s: string[]) { s.lastIndexOf(needle) === s.length - needle.length; }`},

			// lastIndexOf comparing to 0
			{Code: `function f(s: string) { s.lastIndexOf(needle) === 0; }`},

			// match without null check
			{Code: `function f(s: string) { s.match(/^foo/); }`},
			{Code: `function f(s: string) { s.match(/foo$/); }`},
			{Code: `function f(s: string) { s.match(/^foo/) + 1; }`},
			{Code: `function f(s: string) { s.match(/foo$/) + 1; }`},

			// match on non-string type
			{Code: `function f(s: { match(x: any): boolean }) { s.match(/^foo/) !== null; }`},
			{Code: `function f(s: { match(x: any): boolean }) { s.match(/foo$/) !== null; }`},

			// match with pattern that doesn't have anchor
			{Code: `function f(s: string) { s.match(/foo/) !== null; }`},

			// match with both anchors
			{Code: `function f(s: string) { s.match(/^foo$/) !== null; }`},

			// match with other regex features
			{Code: `function f(s: string) { s.match(/^foo./) !== null; }`},
			{Code: `function f(s: string) { s.match(/^foo|bar/) !== null; }`},

			// match with non-literal regex
			{Code: `function f(s: string) { s.match(new RegExp('')) !== null; }`},
			{Code: `function f(s: string) { s.match(pattern) !== null; }`},
			{Code: `function f(s: string) { s.match(new RegExp('^/!{[', 'u')) !== null; }`},

			// match without arguments
			{Code: `function f(s: string) { s.match() !== null; }`},
			{Code: `function f(s: string) { s.match(777) !== null; }`},

			// slice on arrays
			{Code: `function f(s: string[]) { s.slice(0, needle.length) === needle; }`},
			{Code: `function f(s: string[]) { s.slice(-needle.length) === needle; }`},

			// slice with non-0 start
			{Code: `function f(s: string) { s.slice(1, 4) === 'bar'; }`},
			{Code: `function f(s: string) { s.slice(-4, -1) === 'bar'; }`},
			{Code: `function f(s: string) { s.slice(1) === 'bar'; }`},
			{Code: `function f(s: string | null) { s?.slice(1) === 'bar'; }`},

			// slice with non-literal length
			{Code: `function f(s: string) { s.slice(0, -4) === 'car'; }`},
			{Code: `function f(x: string, s: string) { x.endsWith('foo') && x.slice(0, -4) === 'bar'; }`},
			{Code: `function f(s: string) { s.slice(0, length) === needle; }`},
			{Code: `function f(s: string) { s.slice(-length) === needle; }`},
			{Code: `function f(s: string) { s.slice(0, 3) === needle; }`},

			// RegExp.test with non-literal pattern
			{Code: `function f(s: string) { pattern.test(s); }`},
			{Code: `function f(s: string) { /^bar/.test(); }`},
			{Code: `function f(x: { test(): void }, s: string) { x.test(s); }`},

			// allowSingleElementEquality option
			{
				Code:    `declare const s: string; s[0] === 'a';`,
				Options: []interface{}{map[string]interface{}{"allowSingleElementEquality": "always"}},
			},
			{
				Code:    `declare const s: string; s[s.length - 1] === 'a';`,
				Options: []interface{}{map[string]interface{}{"allowSingleElementEquality": "always"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// String indexing - s[0] === 'a'
			{
				Code: `function f(s: string) { s[0] === 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { s?.[0] === 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s?.startsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { s[0] !== 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.startsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { s?.[0] !== 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s?.startsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { s[0] == 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { s[0] != 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.startsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { s[0] === 'あ'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith('あ'); }`},
			},
			{
				Code: `function f(s: string) { s[0] === '👍'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
			},
			{
				Code: `function f(s: string, t: string) { s[0] === t; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 36},
				},
			},
			{
				Code: `function f(s: string) { s[s.length - 1] === 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { (s)[0] === ("a") }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { (s).startsWith("a") }`},
			},

			// String#charAt
			{
				Code: `function f(s: string) { s.charAt(0) === 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { s.charAt(0) !== 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.startsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { s.charAt(0) == 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { s.charAt(0) != 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.startsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { s.charAt(0) === 'あ'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith('あ'); }`},
			},
			{
				Code: `function f(s: string) { s.charAt(0) === '👍'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
			},
			{
				Code: `function f(s: string, t: string) { s.charAt(0) === t; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 36},
				},
			},
			{
				Code: `function f(s: string) { s.charAt(s.length - 1) === 'a'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith('a'); }`},
			},
			{
				Code: `function f(s: string) { (s).charAt(0) === "a"; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { (s).startsWith("a"); }`},
			},

			// String#indexOf
			{
				Code: `function f(s: string) { s.indexOf(needle) === 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith(needle); }`},
			},
			{
				Code: `function f(s: string) { s?.indexOf(needle) === 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s?.startsWith(needle); }`},
			},
			{
				Code: `function f(s: string) { s.indexOf(needle) !== 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.startsWith(needle); }`},
			},
			{
				Code: `function f(s: string) { s.indexOf(needle) == 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith(needle); }`},
			},
			{
				Code: `function f(s: string) { s.indexOf(needle) != 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.startsWith(needle); }`},
			},

			// String#lastIndexOf
			{
				Code: `function f(s: string) { s.lastIndexOf('bar') === s.length - 3; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.lastIndexOf('bar') !== s.length - 3; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.endsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.lastIndexOf('bar') == s.length - 3; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.lastIndexOf('bar') != s.length - 3; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.endsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.lastIndexOf('bar') === s.length - 'bar'.length; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.lastIndexOf(needle) === s.length - needle.length; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith(needle); }`},
			},

			// String#match
			{
				Code: `function f(s: string) { s.match(/^bar/) !== null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith("bar"); }`},
			},
			{
				Code: `function f(s: string) { s?.match(/^bar/) !== null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s?.startsWith("bar"); }`},
			},
			{
				Code: `function f(s: string) { s.match(/^bar/) != null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith("bar"); }`},
			},
			{
				Code: `function f(s: string) { s.match(/bar$/) !== null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith("bar"); }`},
			},
			{
				Code: `function f(s: string) { s.match(/bar$/) != null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith("bar"); }`},
			},
			{
				Code: `function f(s: string) { s.match(/^bar/) === null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.startsWith("bar"); }`},
			},
			{
				Code: `function f(s: string) { s.match(/^bar/) == null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.startsWith("bar"); }`},
			},
			{
				Code: `function f(s: string) { s.match(/bar$/) === null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.endsWith("bar"); }`},
			},
			{
				Code: `function f(s: string) { s.match(/bar$/) == null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.endsWith("bar"); }`},
			},
			{
				Code: `const pattern = /^bar/; function f(s: string) { s.match(pattern) != null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 50},
				},
				Output: []string{`const pattern = /^bar/; function f(s: string) { s.startsWith("bar"); }`},
			},
			{
				Code: `const pattern = new RegExp('^bar'); function f(s: string) { s.match(pattern) != null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 62},
				},
				Output: []string{`const pattern = new RegExp('^bar'); function f(s: string) { s.startsWith("bar"); }`},
			},
			{
				Code: `const pattern = /^"quoted"/; function f(s: string) { s.match(pattern) != null; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 55},
				},
				Output: []string{`const pattern = /^"quoted"/; function f(s: string) { s.startsWith("\"quoted\""); }`},
			},

			// String#slice
			{
				Code: `function f(s: string) { s.slice(0, 3) === 'bar'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s?.slice(0, 3) === 'bar'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s?.startsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.slice(0, 3) !== 'bar'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.startsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.slice(0, 3) == 'bar'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.slice(0, 3) != 'bar'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.startsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.slice(0, needle.length) === needle; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith(needle); }`},
			},
			{
				Code: `function f(s: string) { s.slice(0, needle.length) == needle; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
			},
			{
				Code: `function f(s: string) { s.slice(-3) === 'bar'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.slice(-3) !== 'bar'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { !s.endsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.slice(-needle.length) === needle; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith(needle); }`},
			},
			{
				Code: `function f(s: string) { s.slice(s.length - needle.length) === needle; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith(needle); }`},
			},
			{
				Code: `function f(s: string) { s.substring(0, 3) === 'bar'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith('bar'); }`},
			},
			{
				Code: `function f(s: string) { s.substring(-3) === 'bar'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
			},
			{
				Code: `function f(s: string) { s.substring(s.length - 3, s.length) === 'bar'; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith('bar'); }`},
			},

			// RegExp#test
			{
				Code: `function f(s: string) { /^bar/.test(s); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.startsWith("bar"); }`},
			},
			{
				Code: `function f(s: string) { /^bar/?.test(s); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s?.startsWith("bar"); }`},
			},
			{
				Code: `function f(s: string) { /bar$/.test(s); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferEndsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { s.endsWith("bar"); }`},
			},
			{
				Code: `const pattern = /^bar/; function f(s: string) { pattern.test(s); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 50},
				},
				Output: []string{`const pattern = /^bar/; function f(s: string) { s.startsWith("bar"); }`},
			},
			{
				Code: `const pattern = new RegExp('^bar'); function f(s: string) { pattern.test(s); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 62},
				},
				Output: []string{`const pattern = new RegExp('^bar'); function f(s: string) { s.startsWith("bar"); }`},
			},
			{
				Code: `const pattern = /^"quoted"/; function f(s: string) { pattern.test(s); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 55},
				},
				Output: []string{`const pattern = /^"quoted"/; function f(s: string) { s.startsWith("\"quoted\""); }`},
			},
			{
				Code: `function f(s: string) { /^bar/.test(a + b); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 25},
				},
				Output: []string{`function f(s: string) { (a + b).startsWith("bar"); }`},
			},

			// Test for variation of string types
			{
				Code: `function f(s: 'a' | 'b') { s.indexOf(needle) === 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 28},
				},
				Output: []string{`function f(s: 'a' | 'b') { s.startsWith(needle); }`},
			},
			{
				Code: `function f<T extends 'a' | 'b'>(s: T) { s.indexOf(needle) === 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 41},
				},
				Output: []string{`function f<T extends 'a' | 'b'>(s: T) { s.startsWith(needle); }`},
			},
			{
				Code: `type SafeString = string & { __HTML_ESCAPED__: void }; function f(s: SafeString) { s.indexOf(needle) === 0; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferStartsWith", Line: 1, Column: 85},
				},
				Output: []string{`type SafeString = string & { __HTML_ESCAPED__: void }; function f(s: SafeString) { s.startsWith(needle); }`},
			},
		},
	)
}
