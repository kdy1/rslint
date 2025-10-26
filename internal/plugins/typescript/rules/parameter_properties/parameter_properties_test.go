package parameter_properties

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestParameterPropertiesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ParameterPropertiesRule,
		// Valid cases (default: prefer class-property)
		[]rule_tester.ValidTestCase{
			// Constructor with undecorated parameter
			{Code: `class Foo { constructor(name: string) {} }`},
			// Constructor with rest parameters
			{Code: `class Foo { constructor(...args: any[]) {} }`},
			// Multiple undecorated parameters
			{Code: `class Foo { constructor(a: string, b: number) {} }`},
			// Constructor overloads
			{Code: `class Foo {
  constructor(a: string);
  constructor(a: string, b: number);
  constructor(a: string, b?: number) {}
}`},
			// Class with no constructor
			{Code: `class Foo { method() {} }`},
			// With allow readonly
			{Code: `class Foo { constructor(readonly name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"readonly"}}},
			// With allow private
			{Code: `class Foo { constructor(private name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"private"}}},
			// With allow protected
			{Code: `class Foo { constructor(protected name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"protected"}}},
			// With allow public
			{Code: `class Foo { constructor(public name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"public"}}},
			// With allow private readonly
			{Code: `class Foo { constructor(private readonly name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"private readonly"}}},
			// With allow protected readonly
			{Code: `class Foo { constructor(protected readonly name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"protected readonly"}}},
			// With allow public readonly
			{Code: `class Foo { constructor(public readonly name: string) {} }`, Options: map[string]interface{}{"allow": []interface{}{"public readonly"}}},
		},
		// Invalid cases (default: prefer class-property)
		[]rule_tester.InvalidTestCase{
			// readonly parameter property
			{
				Code: `class Foo { constructor(readonly name: string) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferClassProperty"},
				},
			},
			// private parameter property
			{
				Code: `class Foo { constructor(private name: string) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferClassProperty"},
				},
			},
			// protected parameter property
			{
				Code: `class Foo { constructor(protected name: string) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferClassProperty"},
				},
			},
			// public parameter property
			{
				Code: `class Foo { constructor(public name: string) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferClassProperty"},
				},
			},
			// private readonly
			{
				Code: `class Foo { constructor(private readonly name: string) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferClassProperty"},
				},
			},
			// protected readonly
			{
				Code: `class Foo { constructor(protected readonly name: string) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferClassProperty"},
				},
			},
			// public readonly
			{
				Code: `class Foo { constructor(public readonly name: string) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferClassProperty"},
				},
			},
			// Multiple parameter properties
			{
				Code: `class Foo { constructor(private a: string, public b: number) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferClassProperty"},
					{MessageId: "preferClassProperty"},
				},
			},
			// Not in allow list (allow private but use public)
			{
				Code: `class Foo { constructor(public name: string) {} }`,
				Options: map[string]interface{}{"allow": []interface{}{"private"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferClassProperty"},
				},
			},
		},
	)
}

func TestParameterPropertiesRulePreferParameterProperty(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ParameterPropertiesRule,
		// Valid cases (prefer parameter-property)
		[]rule_tester.ValidTestCase{
			// Using parameter property
			{Code: `class Foo { constructor(private name: string) {} }`, Options: map[string]interface{}{"prefer": "parameter-property"}},
			// Property without assignment in constructor
			{Code: `class Foo { private name: string; }`, Options: map[string]interface{}{"prefer": "parameter-property"}},
			// Property with initializer
			{Code: `class Foo { private name: string = "default"; constructor(name: string) {} }`, Options: map[string]interface{}{"prefer": "parameter-property"}},
			// Property without constructor parameter
			{Code: `class Foo { private name: string; constructor() { this.name = "default"; } }`, Options: map[string]interface{}{"prefer": "parameter-property"}},
			// Constructor parameter without assignment
			{Code: `class Foo { private name: string; constructor(value: string) {} }`, Options: map[string]interface{}{"prefer": "parameter-property"}},
			// Static property
			{Code: `class Foo { static name: string; constructor(name: string) { Foo.name = name; } }`, Options: map[string]interface{}{"prefer": "parameter-property"}},
			// Allow readonly
			{Code: `class Foo { readonly name: string; constructor(name: string) { this.name = name; } }`, Options: map[string]interface{}{"prefer": "parameter-property", "allow": []interface{}{"readonly"}}},
			// Allow private
			{Code: `class Foo { private name: string; constructor(name: string) { this.name = name; } }`, Options: map[string]interface{}{"prefer": "parameter-property", "allow": []interface{}{"private"}}},
			// Allow public
			{Code: `class Foo { name: string; constructor(name: string) { this.name = name; } }`, Options: map[string]interface{}{"prefer": "parameter-property", "allow": []interface{}{"public"}}},
		},
		// Invalid cases (prefer parameter-property)
		[]rule_tester.InvalidTestCase{
			// Class property assigned in constructor
			{
				Code: `class Foo { private name: string; constructor(name: string) { this.name = name; } }`,
				Options: map[string]interface{}{"prefer": "parameter-property"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferParameterProperty"},
				},
			},
			// Public property (implicit)
			{
				Code: `class Foo { name: string; constructor(name: string) { this.name = name; } }`,
				Options: map[string]interface{}{"prefer": "parameter-property"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferParameterProperty"},
				},
			},
			// Protected property
			{
				Code: `class Foo { protected name: string; constructor(name: string) { this.name = name; } }`,
				Options: map[string]interface{}{"prefer": "parameter-property"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferParameterProperty"},
				},
			},
			// Readonly property
			{
				Code: `class Foo { readonly name: string; constructor(name: string) { this.name = name; } }`,
				Options: map[string]interface{}{"prefer": "parameter-property"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferParameterProperty"},
				},
			},
			// Not in allow list
			{
				Code: `class Foo { name: string; constructor(name: string) { this.name = name; } }`,
				Options: map[string]interface{}{"prefer": "parameter-property", "allow": []interface{}{"private", "protected", "readonly"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferParameterProperty"},
				},
			},
		},
	)
}
