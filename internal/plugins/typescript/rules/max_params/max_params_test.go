package max_params

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMaxParamsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MaxParamsRule, []rule_tester.ValidTestCase{
		// Valid cases from TypeScript-ESLint tests
		{
			Code: `function foo() {}`,
		},
		{
			Code: `const foo = function () {};`,
		},
		{
			Code: `const foo = () => {};`,
		},
		{
			Code: `function foo(a) {}`,
		},
		{
			Code: `
class Foo {
  constructor(a) {}
}
			`,
		},
		{
			Code: `
class Foo {
  method(this: void, a, b, c) {}
}
			`,
		},
		{
			Code: `
class Foo {
  method(this: Foo, a, b) {}
}
			`,
		},
		{
			Code:    `function foo(a, b, c, d) {}`,
			Options: map[string]interface{}{"max": 4},
		},
		{
			Code:    `function foo(a, b, c, d) {}`,
			Options: map[string]interface{}{"maximum": 4},
		},
		{
			Code: `
class Foo {
  method(this: void) {}
}
			`,
			Options: map[string]interface{}{"max": 0},
		},
		{
			Code: `
class Foo {
  method(this: void, a) {}
}
			`,
			Options: map[string]interface{}{"max": 1},
		},
		{
			Code: `
class Foo {
  method(this: void, a) {}
}
			`,
			Options: map[string]interface{}{"countVoidThis": true, "max": 2},
		},
		{
			Code: `
declare function makeDate(m: number, d: number, y: number): Date;
			`,
			Options: map[string]interface{}{"max": 3},
		},
		{
			Code: `
type sum = (a: number, b: number) => number;
			`,
			Options: map[string]interface{}{"max": 2},
		},
	}, []rule_tester.InvalidTestCase{
		// Invalid cases from TypeScript-ESLint tests
		{
			Code: `function foo(a, b, c, d) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "exceed",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `const foo = function (a, b, c, d) {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "exceed",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: `const foo = (a, b, c, d) => {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "exceed",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: `const foo = a => {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "exceed",
					Line:      1,
					Column:    13,
				},
			},
			Options: map[string]interface{}{"max": 0},
		},
		{
			Code: `
class Foo {
  method(this: void, a, b, c, d) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "exceed",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
class Foo {
  method(this: void, a) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "exceed",
					Line:      3,
					Column:    3,
				},
			},
			Options: map[string]interface{}{"countVoidThis": true, "max": 1},
		},
		{
			Code: `
class Foo {
  method(this: Foo, a, b, c) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "exceed",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
declare function makeDate(m: number, d: number, y: number): Date;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "exceed",
					Line:      2,
					Column:    1,
				},
			},
			Options: map[string]interface{}{"max": 1},
		},
		{
			Code: `
type sum = (a: number, b: number) => number;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "exceed",
					Line:      2,
					Column:    12,
				},
			},
			Options: map[string]interface{}{"max": 1},
		},
	})
}
