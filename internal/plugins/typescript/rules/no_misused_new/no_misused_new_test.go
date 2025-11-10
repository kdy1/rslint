package no_misused_new

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMisusedNewRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMisusedNewRule, []rule_tester.ValidTestCase{
		// Valid: abstract class with 'get new()' accessor
		{Code: `
declare abstract class C {
  foo() {}
  get new();
  bar();
}
`},
		// Valid: class with constructor signature
		{Code: `
class C {
  constructor();
}
`},
		// Valid: class expression with constructor signature
		{Code: `
const foo = class {
  constructor();
};
`},
		// Valid: class expression with 'new' method signature (different return type)
		{Code: `
const foo = class {
  new(): X;
};
`},
		// Valid: 'new' method with body is OK
		{Code: `
class C {
  new() {}
}
`},
		// Valid: constructor with body is OK
		{Code: `
class C {
  constructor() {}
}
`},
		// Valid: class expression with 'new' method body
		{Code: `
const foo = class {
  new() {}
};
`},
		// Valid: class expression with constructor body
		{Code: `
const foo = class {
  constructor() {}
};
`},
		// Valid: interface 'new' with different return type
		{Code: `
interface I {
  new (): {};
}
`},
		// Valid: 'new' OK in type literal (we don't know the type name)
		{Code: `type T = { new (): T };`},
		// Valid: default export class with constructor
		{Code: `
export default class {
  constructor();
}
`},
		// Valid: interface 'new' with generic returning different type
		{Code: `
interface foo {
  new <T>(): bar<T>;
}
`},
		// Valid: interface 'new' with generic returning string literal
		{Code: `
interface foo {
  new <T>(): 'x';
}
`},
	}, []rule_tester.InvalidTestCase{
		// Invalid: interface with both 'new' returning same type and 'constructor'
		{
			Code: `
interface I {
  new (): I;
  constructor(): void;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageInterface",
					Line:      3,
					Column:    3,
				},
				{
					MessageId: "errorMessageInterface",
					Line:      4,
					Column:    3,
				},
			},
		},
		// Invalid: generic interface with 'new' returning same generic type
		{
			Code: `
interface G {
  new <T>(): G<T>;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageInterface",
					Line:      3,
					Column:    3,
				},
			},
		},
		// Invalid: type literal with 'constructor'
		{
			Code: `
type T = {
  constructor(): void;
};
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageInterface",
					Line:      3,
					Column:    3,
				},
			},
		},
		// Invalid: class with 'new' method signature returning same type
		{
			Code: `
class C {
  new(): C;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageClass",
					Line:      3,
					Column:    3,
				},
			},
		},
		// Invalid: abstract class with 'new' method signature returning same type
		{
			Code: `
declare abstract class C {
  new(): C;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageClass",
					Line:      3,
					Column:    3,
				},
			},
		},
		// Invalid: interface with 'constructor' returning string
		{
			Code: `
interface I {
  constructor(): '';
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorMessageInterface",
					Line:      3,
					Column:    3,
				},
			},
		},
	})
}
