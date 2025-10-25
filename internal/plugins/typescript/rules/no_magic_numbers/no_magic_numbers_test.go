package no_magic_numbers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoMagicNumbersRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMagicNumbersRule, []rule_tester.ValidTestCase{
		// Valid: Numbers in ignore list
		{
			Code:    `const foo = 42;`,
			Options: map[string]interface{}{"ignore": []interface{}{42.0}},
		},
		{
			Code:    `const foo = -1;`,
			Options: map[string]interface{}{"ignore": []interface{}{-1.0}},
		},
		{
			Code:    `const foo = 0;`,
			Options: map[string]interface{}{"ignore": []interface{}{0.0}},
		},
		{
			Code:    `const foo = 1;`,
			Options: map[string]interface{}{"ignore": []interface{}{1.0}},
		},

		// Valid: ignoreEnums option
		{
			Code: `
enum Foo {
  A = 1,
  B = 2,
  C = 3
}
`,
			Options: map[string]interface{}{"ignoreEnums": true},
		},
		{
			Code: `
enum StatusCode {
  OK = 200,
  NotFound = 404,
  ServerError = 500
}
`,
			Options: map[string]interface{}{"ignoreEnums": true},
		},

		// Valid: ignoreNumericLiteralTypes option
		{
			Code:    `type Foo = 1 | 2 | 3;`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": true},
		},
		{
			Code:    `type Port = 8080;`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": true},
		},

		// Valid: ignoreReadonlyClassProperties option
		{
			Code: `
class Foo {
  readonly bar = 1;
  readonly baz = 2;
}
`,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": true},
		},
		{
			Code: `
class Config {
  private readonly timeout = 5000;
  public readonly maxRetries = 3;
}
`,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": true},
		},

		// Valid: ignoreTypeIndexes option
		{
			Code:    `type Foo<T> = T[0];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": true},
		},
		{
			Code:    `type First<T extends readonly unknown[]> = T[0];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": true},
		},

		// Valid: ignoreArrayIndexes option
		{
			Code:    `const foo = arr[0];`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
		},
		{
			Code:    `const bar = items[1];`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
		},
		{
			Code:    `const x = data[42];`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
		},

		// Valid: ignoreDefaultValues option
		{
			Code:    `function foo(x = 5) { return x; }`,
			Options: map[string]interface{}{"ignoreDefaultValues": true},
		},
		{
			Code:    `const bar = (y = 10) => y;`,
			Options: map[string]interface{}{"ignoreDefaultValues": true},
		},
		{
			Code:    `const { a = 3 } = obj;`,
			Options: map[string]interface{}{"ignoreDefaultValues": true},
		},

		// Valid: ignoreClassFieldInitialValues option
		{
			Code: `
class Foo {
  bar = 1;
  baz = 2;
}
`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": true},
		},
		{
			Code: `
class Config {
  private timeout = 5000;
  static maxRetries = 3;
}
`,
			Options: map[string]interface{}{"ignoreClassFieldInitialValues": true},
		},

		// Valid: detectObjects disabled (default)
		{
			Code: `const obj = { a: 1, b: 2 };`,
		},
		{
			Code: `const config = { port: 8080, timeout: 5000 };`,
		},

		// Valid: enforceConst with const declaration
		{
			Code:    `const FOO = 42;`,
			Options: map[string]interface{}{"enforceConst": true},
		},

		// Valid: combination of options
		{
			Code: `
enum Status {
  Active = 1,
  Inactive = 0
}

type Port = 8080;

class Config {
  readonly timeout = 5000;
}

function foo(x = 10) {
  return x;
}
`,
			Options: map[string]interface{}{
				"ignoreEnums":                   true,
				"ignoreNumericLiteralTypes":     true,
				"ignoreReadonlyClassProperties": true,
				"ignoreDefaultValues":           true,
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Invalid: Basic magic number
		{
			Code: `const foo = 42;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: `const bar = 3.14;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    13,
				},
			},
		},

		// Invalid: Magic number in variable declaration
		{
			Code: `let x = 100;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    9,
				},
			},
		},

		// Invalid: Magic number in expression
		{
			Code: `const result = value * 2;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    24,
				},
			},
		},

		// Invalid: Enum without ignoreEnums
		{
			Code: `
enum Status {
  Active = 1,
  Inactive = 0
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      3,
					Column:    12,
				},
				{
					MessageId: "noMagic",
					Line:      4,
					Column:    14,
				},
			},
		},

		// Invalid: Numeric literal type without ignoreNumericLiteralTypes
		{
			Code: `type Port = 8080;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    13,
				},
			},
		},

		// Invalid: Readonly class property without ignoreReadonlyClassProperties
		{
			Code: `
class Config {
  readonly timeout = 5000;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      3,
					Column:    22,
				},
			},
		},

		// Invalid: Type index without ignoreTypeIndexes
		{
			Code: `type First<T extends readonly unknown[]> = T[0];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    46,
				},
			},
		},

		// Invalid: Array index without ignoreArrayIndexes
		{
			Code: `const first = arr[0];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    19,
				},
			},
		},

		// Invalid: Default value without ignoreDefaultValues
		{
			Code: `function foo(x = 5) { return x; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    18,
				},
			},
		},

		// Invalid: Class field initial value without ignoreClassFieldInitialValues
		{
			Code: `
class Foo {
  bar = 1;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      3,
					Column:    9,
				},
			},
		},

		// Invalid: detectObjects enabled
		{
			Code: `const obj = { a: 1, b: 2 };`,
			Options: map[string]interface{}{"detectObjects": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    18,
				},
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    24,
				},
			},
		},

		// Invalid: enforceConst with let declaration
		{
			Code:    `let FOO = 42;`,
			Options: map[string]interface{}{"enforceConst": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    11,
				},
			},
		},

		// Invalid: enforceConst with var declaration
		{
			Code:    `var BAR = 100;`,
			Options: map[string]interface{}{"enforceConst": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    11,
				},
			},
		},

		// Invalid: Multiple magic numbers
		{
			Code: `
const a = 1;
const b = 2;
const c = 3;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      2,
					Column:    11,
				},
				{
					MessageId: "noMagic",
					Line:      3,
					Column:    11,
				},
				{
					MessageId: "noMagic",
					Line:      4,
					Column:    11,
				},
			},
		},

		// Invalid: Magic number not in ignore list
		{
			Code:    `const foo = 42;`,
			Options: map[string]interface{}{"ignore": []interface{}{1.0, 2.0, 3.0}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    13,
				},
			},
		},

		// Invalid: Regular class property (not readonly)
		{
			Code: `
class Foo {
  bar = 1;
}
`,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      3,
					Column:    9,
				},
			},
		},

		// Invalid: Combination test - some options enabled
		{
			Code: `
enum Status {
  Active = 1,
  Inactive = 0
}

type Port = 8080;

class Config {
  readonly timeout = 5000;
  retries = 3;
}
`,
			Options: map[string]interface{}{
				"ignoreEnums":                   true,
				"ignoreReadonlyClassProperties": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      7,
					Column:    13,
				},
				{
					MessageId: "noMagic",
					Line:      11,
					Column:    13,
				},
			},
		},
	})
}

// Test specific TypeScript-only features
func TestNoMagicNumbersTypeScriptFeatures(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMagicNumbersRule, []rule_tester.ValidTestCase{
		// Valid: Union type with ignoreNumericLiteralTypes
		{
			Code:    `type Status = 1 | 2 | 3;`,
			Options: map[string]interface{}{"ignoreNumericLiteralTypes": true},
		},

		// Valid: Type parameter with ignoreTypeIndexes
		{
			Code:    `type Second<T extends readonly unknown[]> = T[1];`,
			Options: map[string]interface{}{"ignoreTypeIndexes": true},
		},

		// Valid: Private readonly property
		{
			Code: `
class Foo {
  private readonly bar = 100;
}
`,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": true},
		},

		// Valid: Static readonly property
		{
			Code: `
class Foo {
  static readonly bar = 100;
}
`,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": true},
		},
	}, []rule_tester.InvalidTestCase{
		// Invalid: Union type without ignoreNumericLiteralTypes
		{
			Code: `type Status = 1 | 2 | 3;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    15,
				},
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    19,
				},
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    23,
				},
			},
		},

		// Invalid: Private non-readonly property
		{
			Code: `
class Foo {
  private bar = 100;
}
`,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      3,
					Column:    17,
				},
			},
		},

		// Invalid: Static non-readonly property
		{
			Code: `
class Foo {
  static bar = 100;
}
`,
			Options: map[string]interface{}{"ignoreReadonlyClassProperties": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      3,
					Column:    16,
				},
			},
		},
	})
}

// Test edge cases
func TestNoMagicNumbersEdgeCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoMagicNumbersRule, []rule_tester.ValidTestCase{
		// Valid: Negative numbers in ignore list
		{
			Code:    `const foo = -1;`,
			Options: map[string]interface{}{"ignore": []interface{}{-1.0}},
		},

		// Valid: Decimal numbers in ignore list
		{
			Code:    `const pi = 3.14;`,
			Options: map[string]interface{}{"ignore": []interface{}{3.14}},
		},

		// Valid: Multiple ignore values
		{
			Code:    `const a = 1; const b = 2; const c = 3;`,
			Options: map[string]interface{}{"ignore": []interface{}{1.0, 2.0, 3.0}},
		},

		// Valid: Destructuring default with ignoreDefaultValues
		{
			Code:    `const { x = 10, y = 20 } = obj;`,
			Options: map[string]interface{}{"ignoreDefaultValues": true},
		},

		// Valid: Arrow function default parameter (only default value is ignored, not function body)
		{
			Code:    `const foo = (x = 5) => x;`,
			Options: map[string]interface{}{"ignoreDefaultValues": true},
		},
	}, []rule_tester.InvalidTestCase{
		// Invalid: Array index out of valid range (>= 2^32)
		{
			Code:    `const foo = arr[4294967295];`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    17,
				},
			},
		},

		// Invalid: Negative array index
		{
			Code:    `const foo = arr[-1];`,
			Options: map[string]interface{}{"ignoreArrayIndexes": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noMagic",
					Line:      1,
					Column:    18,
				},
			},
		},
	})
}
