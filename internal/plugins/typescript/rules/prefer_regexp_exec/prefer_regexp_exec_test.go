package prefer_regexp_exec

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferRegexpExecRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferRegexpExecRule, []rule_tester.ValidTestCase{
		// No arguments to match
		{Code: `'something'.match();`},

		// Global flag present
		{Code: `'something'.match(/thing/g);`},

		// Variable with global flag
		{Code: `
const text = 'something';
const search = /thing/g;
text.match(search);
		`},

		// match method with different signature
		{Code: `
const match = (s: RegExp) => 'something';
match(/thing/);
		`},

		// Object method with name match
		{Code: `
const a = { match: (s: RegExp) => 'something' };
a.match(/thing/);
		`},

		// Union type with string array
		{Code: `
function f(s: string | string[]) {
  s.match(/e/);
}
		`},

		// Non-RegExp arguments (number literals)
		{Code: `(Math.random() > 0.5 ? 'abc' : 123).match(2);`},
		{Code: `'212'.match(2);`},
		{Code: `'212'.match(+2);`},
		{Code: `'oNaNo'.match(NaN);`},
		{Code: `'Infinity contains -Infinity and +Infinity in JavaScript.'.match(Infinity);`},
		{Code: `'Infinity contains -Infinity and +Infinity in JavaScript.'.match(+Infinity);`},
		{Code: `'Infinity contains -Infinity and +Infinity in JavaScript.'.match(-Infinity);`},
		{Code: `'void and null'.match(null);`},

		// Mixed array of matchers
		{Code: `
const matchers = ['package-lock.json', /regexp/];
const file = '';
matchers.some(matcher => !!file.match(matcher));
		`},

		{Code: `
const matchers = [/regexp/, 'package-lock.json'];
const file = '';
matchers.some(matcher => !!file.match(matcher));
		`},

		{Code: `
const matchers = [{ match: (s: RegExp) => false }];
const file = '';
matchers.some(matcher => !!file.match(matcher));
		`},

		// RegExp with global flag - issue #3477
		{Code: `
function test(pattern: string) {
  'hello hello'.match(RegExp(pattern, 'g'))?.reduce(() => []);
}
		`},

		{Code: `
function test(pattern: string) {
  'hello hello'.match(new RegExp(pattern, 'gi'))?.reduce(() => []);
}
		`},

		{Code: `
const matchCount = (str: string, re: RegExp) => {
  return (str.match(re) || []).length;
};
		`},

		// Invalid regex pattern - issue #6928
		{Code: `
function test(str: string) {
  str.match('[a-z');
}
		`},

		// Declared variable with RegExp type
		{Code: `
const text = 'something';
declare const search: RegExp;
text.match(search);
		`},

		// Property access returning RegExp - issue #8614
		{Code: `
const text = 'something';
declare const obj: { search: RegExp };
text.match(obj.search);
		`},

		// Function returning RegExp
		{Code: `
const text = 'something';
declare function returnsRegexp(): RegExp;
text.match(returnsRegexp());
		`},
	}, []rule_tester.InvalidTestCase{
		// Literal RegExp without global flag
		{
			Code: `'something'.match(/thing/);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regExpExecOverStringMatch",
					Line:      1,
					Column:    13,
				},
			},
			Output: []string{`/thing/.exec('something');`},
		},

		// String literal argument
		{
			Code: `'something'.match('^[a-z]+thing/?$');`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regExpExecOverStringMatch",
					Line:      1,
					Column:    13,
				},
			},
			Output: []string{`/^[a-z]+thing\/?$/.exec('something');`},
		},

		// RegExp variable
		{
			Code: `
const text = 'something';
const search = /thing/;
text.match(search);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regExpExecOverStringMatch",
					Line:      4,
					Column:    6,
				},
			},
			Output: []string{`
const text = 'something';
const search = /thing/;
search.exec(text);
			`},
		},

		// String variable
		{
			Code: `
const text = 'something';
const search = 'thing';
text.match(search);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regExpExecOverStringMatch",
					Line:      4,
					Column:    6,
				},
			},
			Output: []string{`
const text = 'something';
const search = 'thing';
RegExp(search).exec(text);
			`},
		},

		// String literal type
		{
			Code: `
function f(s: 'a' | 'b') {
  s.match('a');
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regExpExecOverStringMatch",
					Line:      3,
					Column:    5,
				},
			},
			Output: []string{`
function f(s: 'a' | 'b') {
  /a/.exec(s);
}
			`},
		},

		// Intersection type with string
		{
			Code: `
type SafeString = string & { __HTML_ESCAPED__: void };
function f(s: SafeString) {
  s.match(/thing/);
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regExpExecOverStringMatch",
					Line:      4,
					Column:    5,
				},
			},
			Output: []string{`
type SafeString = string & { __HTML_ESCAPED__: void };
function f(s: SafeString) {
  /thing/.exec(s);
}
			`},
		},

		// Generic type constraint
		{
			Code: `
function f<T extends 'a' | 'b'>(s: T) {
  s.match(/thing/);
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regExpExecOverStringMatch",
					Line:      3,
					Column:    5,
				},
			},
			Output: []string{`
function f<T extends 'a' | 'b'>(s: T) {
  /thing/.exec(s);
}
			`},
		},

		// new RegExp without global flag
		{
			Code: `
const text = 'something';
const search = new RegExp('test', '');
text.match(search);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regExpExecOverStringMatch",
					Line:      4,
					Column:    6,
				},
			},
			Output: []string{`
const text = 'something';
const search = new RegExp('test', '');
search.exec(text);
			`},
		},

		// new RegExp with undefined flags
		{
			Code: `
function test(pattern: string) {
  'check'.match(new RegExp(pattern, undefined));
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regExpExecOverStringMatch",
					Line:      3,
					Column:    11,
				},
			},
			Output: []string{`
function test(pattern: string) {
  new RegExp(pattern, undefined).exec('check');
}
			`},
		},

		// new RegExp with template literals - issue #3941
		{
			Code: "function temp(text: string): void {\n  text.match(new RegExp(`${'hello'}`));\n  text.match(new RegExp(`${'hello'.toString()}`));\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "regExpExecOverStringMatch",
					Line:      2,
					Column:    8,
				},
				{
					MessageId: "regExpExecOverStringMatch",
					Line:      3,
					Column:    8,
				},
			},
			Output: []string{"function temp(text: string): void {\n  new RegExp(`${'hello'}`).exec(text);\n  new RegExp(`${'hello'.toString()}`).exec(text);\n}"},
		},
	})
}
