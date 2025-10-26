package no_empty_object_type

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoEmptyObjectTypeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyObjectTypeRule,
		[]rule_tester.ValidTestCase{
			// Interface with properties
			{Code: `interface Base { name: string; }`},

			// Interface extending multiple types
			{Code: `
interface Base { name: string; }
interface Derived { age: number; }
interface Both extends Base, Derived {}
`},

			// Empty interface with 'always' option
			{
				Code:    `interface Base {}`,
				Options: map[string]any{"allowInterfaces": "always"},
			},

			// Empty interface with single extends and appropriate option
			{
				Code: `
interface Base { name: string; }
interface Derived extends Base {}
`,
				Options: map[string]any{"allowInterfaces": "with-single-extends"},
			},

			// Object type declarations
			{Code: `let value: object;`},
			{Code: `let value: Object;`},
			{Code: `let value: { inner: true };`},

			// Generic type with empty object (intersection)
			{Code: `type MyNonNullable<T> = T & {};`},

			// Empty type alias with 'always' option
			{
				Code:    `type Base = {};`,
				Options: map[string]any{"allowObjectTypes": "always"},
			},

			// Matching allowWithName patterns
			{
				Code:    `type Base = {};`,
				Options: map[string]any{"allowWithName": "Base"},
			},
			{
				Code:    `type BaseProps = {};`,
				Options: map[string]any{"allowWithName": "Props$"},
			},
			{
				Code:    `interface Base {}`,
				Options: map[string]any{"allowWithName": "Base"},
			},
			{
				Code:    `interface BaseProps {}`,
				Options: map[string]any{"allowWithName": "Props$"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Empty interface (default)
			{
				Code: `interface Base {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyInterface",
					},
				},
				// Note: Suggestions are provided but Output is not mandatory in the test framework
			},

			// Empty interface with explicit 'never' option
			{
				Code: `interface Base {}`,
				Options: map[string]any{"allowInterfaces": "never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyInterface",
					},
				},
			},

			// Empty interface extending single interface
			{
				Code: `
interface Base { name: string; }
interface Derived extends Base {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyInterfaceWithSuper",
						Line:      3,
					},
				},
			},

			// Empty interface extending generic types
			{
				Code: `interface Base extends Array<number> {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyInterfaceWithSuper",
					},
				},
			},

			// Empty type alias (default)
			{
				Code: `type Base = {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyObject",
					},
				},
			},

			// Empty type in variable annotation
			{
				Code: `let value: {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyObject",
					},
				},
			},

			// Multiline empty object
			{
				Code: `let value: {
  /* empty */
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyObject",
					},
				},
			},

			// Empty object in union type
			{
				Code: `type MyUnion<T> = T | {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyObject",
					},
				},
			},

			// Mismatched allowWithName (empty object not matching pattern)
			{
				Code:    `type Base = {} | null;`,
				Options: map[string]any{"allowWithName": "Base"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyObject",
					},
				},
			},

			// Empty interface not matching allowWithName pattern
			{
				Code:    `interface Base {}`,
				Options: map[string]any{"allowWithName": ".*Props$"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyInterface",
					},
				},
			},

			// Multiple errors in one file
			{
				Code: `
interface Foo {}
interface Bar extends Foo {}
type Baz = {};
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyInterface",
						Line:      2,
					},
					{
						MessageId: "noEmptyInterfaceWithSuper",
						Line:      3,
					},
					{
						MessageId: "noEmptyObject",
						Line:      4,
					},
				},
			},

			// Empty interface with type parameters
			{
				Code: `interface Base<T> {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyInterface",
					},
				},
			},

			// Empty interface extending with type parameters
			{
				Code: `
interface Base<T> { value: T; }
interface Derived<T> extends Base<T> {}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyInterfaceWithSuper",
						Line:      3,
					},
				},
			},

			// Function parameter with empty object type
			{
				Code: `function foo(obj: {}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyObject",
					},
				},
			},

			// Return type with empty object
			{
				Code: `function foo(): {} { return {}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyObject",
					},
				},
			},

			// Array type with empty object
			{
				Code: `type Foo = Array<{}>;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyObject",
					},
				},
			},

			// Nested empty object
			{
				Code: `interface Base extends Array<number | {}> {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noEmptyInterfaceWithSuper",
					},
					{
						MessageId: "noEmptyObject",
					},
				},
			},
		},
	)
}
