package parameter_properties

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestParameterPropertiesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ParameterPropertiesRule, []rule_tester.ValidTestCase{
		// Default (prefer: "class-property") - Valid cases
		{
			Code: `
class Foo {
  constructor(name: string) {}
}
`,
		},
		{
			Code: `
class Foo {
  constructor(name: string);
  constructor(name: string, age: number) {}
}
`,
		},
		{
			Code: `
class Foo {
  name: string;
  constructor(name: string) {
    this.name = name;
  }
}
`,
		},
		{
			Code: `
class Foo {
  private name: string;
  constructor(name: string) {
    this.name = name;
  }
}
`,
		},
		// Allow readonly parameter properties
		{
			Code: `
class Foo {
  constructor(readonly name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "class-property",
				"allow":  []interface{}{"readonly"},
			},
		},
		// Allow private parameter properties
		{
			Code: `
class Foo {
  constructor(private name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "class-property",
				"allow":  []interface{}{"private"},
			},
		},
		// Allow protected parameter properties
		{
			Code: `
class Foo {
  constructor(protected name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "class-property",
				"allow":  []interface{}{"protected"},
			},
		},
		// Allow public parameter properties
		{
			Code: `
class Foo {
  constructor(public name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "class-property",
				"allow":  []interface{}{"public"},
			},
		},
		// Allow private readonly parameter properties
		{
			Code: `
class Foo {
  constructor(private readonly name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "class-property",
				"allow":  []interface{}{"private readonly"},
			},
		},
		// Allow protected readonly parameter properties
		{
			Code: `
class Foo {
  constructor(protected readonly name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "class-property",
				"allow":  []interface{}{"protected readonly"},
			},
		},
		// Allow public readonly parameter properties
		{
			Code: `
class Foo {
  constructor(public readonly name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "class-property",
				"allow":  []interface{}{"public readonly"},
			},
		},
		// Allow multiple types
		{
			Code: `
class Foo {
  constructor(
    private name: string,
    protected readonly age: number,
  ) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "class-property",
				"allow":  []interface{}{"private", "protected readonly"},
			},
		},
		// prefer: "parameter-property" - Valid cases
		{
			Code: `
class Foo {
  constructor(public name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
		},
		{
			Code: `
class Foo {
  constructor(private name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
		},
		{
			Code: `
class Foo {
  constructor(protected name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
		},
		{
			Code: `
class Foo {
  constructor(readonly name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
		},
		{
			Code: `
class Foo {
  constructor(private readonly name: string) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
		},
		// Properties with initializers should not be reported
		{
			Code: `
class Foo {
  private name: string = "default";
  constructor(name: string) {
    this.name = name;
  }
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
		},
		// Properties not assigned from constructor parameters should not be reported
		{
			Code: `
class Foo {
  private name: string;
  constructor() {
    this.name = "default";
  }
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
		},
		// Properties without modifiers should not be reported
		{
			Code: `
class Foo {
  name: string;
  constructor(name: string) {
    this.name = name;
  }
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Default (prefer: "class-property") - Invalid cases
		{
			Code: `
class Foo {
  constructor(public name: string) {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferClassProperty",
					Line:      3,
					Column:    22,
				},
			},
		},
		{
			Code: `
class Foo {
  constructor(private name: string) {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferClassProperty",
					Line:      3,
					Column:    23,
				},
			},
		},
		{
			Code: `
class Foo {
  constructor(protected name: string) {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferClassProperty",
					Line:      3,
					Column:    25,
				},
			},
		},
		{
			Code: `
class Foo {
  constructor(readonly name: string) {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferClassProperty",
					Line:      3,
					Column:    24,
				},
			},
		},
		{
			Code: `
class Foo {
  constructor(private readonly name: string) {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferClassProperty",
					Line:      3,
					Column:    32,
				},
			},
		},
		{
			Code: `
class Foo {
  constructor(
    public name: string,
    private age: number,
  ) {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferClassProperty",
					Line:      4,
					Column:    12,
				},
				{
					MessageId: "preferClassProperty",
					Line:      5,
					Column:    13,
				},
			},
		},
		// With allow option - should still report non-allowed modifiers
		{
			Code: `
class Foo {
  constructor(
    private name: string,
    public age: number,
  ) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "class-property",
				"allow":  []interface{}{"private"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferClassProperty",
					Line:      5,
					Column:    12,
				},
			},
		},
		{
			Code: `
class Foo {
  constructor(
    private readonly name: string,
    protected readonly age: number,
  ) {}
}
`,
			Options: map[string]interface{}{
				"prefer": "class-property",
				"allow":  []interface{}{"private readonly"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferClassProperty",
					Line:      5,
					Column:    24,
				},
			},
		},
		// prefer: "parameter-property" - Invalid cases
		{
			Code: `
class Foo {
  private name: string;
  constructor(name: string) {
    this.name = name;
  }
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferParameterProperty",
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code: `
class Foo {
  public name: string;
  constructor(name: string) {
    this.name = name;
  }
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferParameterProperty",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
class Foo {
  protected name: string;
  constructor(name: string) {
    this.name = name;
  }
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferParameterProperty",
					Line:      3,
					Column:    13,
				},
			},
		},
		{
			Code: `
class Foo {
  readonly name: string;
  constructor(name: string) {
    this.name = name;
  }
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferParameterProperty",
					Line:      3,
					Column:    12,
				},
			},
		},
		{
			Code: `
class Foo {
  private readonly name: string;
  constructor(name: string) {
    this.name = name;
  }
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferParameterProperty",
					Line:      3,
					Column:    20,
				},
			},
		},
		{
			Code: `
class Foo {
  private name: string;
  protected age: number;
  constructor(name: string, age: number) {
    this.name = name;
    this.age = age;
  }
}
`,
			Options: map[string]interface{}{
				"prefer": "parameter-property",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferParameterProperty",
					Line:      3,
					Column:    11,
				},
				{
					MessageId: "preferParameterProperty",
					Line:      4,
					Column:    13,
				},
			},
		},
	})
}
