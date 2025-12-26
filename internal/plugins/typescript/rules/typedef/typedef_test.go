package typedef

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestTypedefRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &TypedefRule, []rule_tester.ValidTestCase{
		// Array destructuring
		{
			Code: `function foo(...[a]: string[]) {}`,
			Options: []interface{}{map[string]interface{}{
				"arrayDestructuring": true,
			}},
		},
		{
			Code: `const foo = (...[a]: string[]) => {};`,
			Options: []interface{}{map[string]interface{}{
				"arrayDestructuring": true,
			}},
		},
		{
			Code: `const [a]: [number] = [1];`,
			Options: []interface{}{map[string]interface{}{
				"arrayDestructuring": true,
			}},
		},
		{
			Code: `const [a, b]: [number, number] = [1, 2];`,
			Options: []interface{}{map[string]interface{}{
				"arrayDestructuring": true,
			}},
		},
		{
			Code: `[a] = [1];`,
			Options: []interface{}{map[string]interface{}{
				"arrayDestructuring": true,
			}},
		},
		{
			Code: `const [[a]]: number[][] = [[1]];`,
			Options: []interface{}{map[string]interface{}{
				"arrayDestructuring": true,
			}},
		},
		{
			Code: `const [a] = [1];`,
			Options: []interface{}{map[string]interface{}{
				"arrayDestructuring": false,
			}},
		},
		{
			Code: `for (const [key, val] of new Map([['key', 1]])) {}`,
			Options: []interface{}{map[string]interface{}{
				"arrayDestructuring": true,
			}},
		},
		{
			Code: `for (const [[key]] of [[['key']]]) {}`,
			Options: []interface{}{map[string]interface{}{
				"arrayDestructuring": true,
			}},
		},
		{
			Code: `let a: number; [a] = [1];`,
		},

		// Arrow parameters
		{Code: `((a: number): void => {})();`},
		{Code: `((a: string, b: string): void => {})();`},
		{
			Code: `((a: number): void => {})();`,
			Options: []interface{}{map[string]interface{}{
				"arrowParameter": false,
			}},
		},
		{
			Code: `((a: string, b: string): void => {})();`,
			Options: []interface{}{map[string]interface{}{
				"arrowParameter": false,
			}},
		},

		// Member variable declarations
		{Code: `class Test { state: number; }`},
		{Code: `class Test { state: number = 1; }`},
		{
			Code: `class Test { state = 1; }`,
			Options: []interface{}{map[string]interface{}{
				"memberVariableDeclaration": false,
			}},
		},

		// Object destructuring
		{
			Code: `const { a }: { a: number } = { a: 1 };`,
			Options: []interface{}{map[string]interface{}{
				"objectDestructuring": true,
			}},
		},
		{
			Code: `const { a, b }: { [i: string]: number } = { a: 1, b: 2 };`,
			Options: []interface{}{map[string]interface{}{
				"objectDestructuring": true,
			}},
		},
		{
			Code: `for (const { p1: { p2: { p3 } } } of [{ p1: { p2: { p3: 'value' } } }]) {}`,
			Options: []interface{}{map[string]interface{}{
				"objectDestructuring": true,
			}},
		},
		{
			Code: `const { a } = { a: 1 };`,
			Options: []interface{}{map[string]interface{}{
				"objectDestructuring": false,
			}},
		},
		{
			Code: `for (const { key, val } of [{ key: 'key', val: 1 }]) {}`,
			Options: []interface{}{map[string]interface{}{
				"objectDestructuring": true,
			}},
		},

		// Function parameters
		{Code: `function receivesNumber(a: number): void {}`},
		{Code: `function receivesStrings(a: string, b: string): void {}`},
		{Code: `function receivesNumber([a]: [number]): void {}`},
		{Code: `function receivesNumbers([a, b]: number[]): void {}`},
		{Code: `function receivesString({ a }: { a: string }): void {}`},
		{Code: `function receivesStrings({ a, b }: { [i: string]: string }): void {}`},
		{Code: `function receivesNumber(a: number = 123): void {}`},

		// Constructor parameters
		{Code: `class Test { constructor() {} }`},
		{Code: `class Test { constructor(param: string) {} }`},
		{Code: `class Test { constructor(param: string = 'something') {} }`},
		{Code: `class Test { constructor(private param: string = 'something') {} }`},

		// Method parameters
		{Code: `class Test { public method(x: number): number { return x; } }`},

		// Property declarations (interface/type)
		{
			Code: `interface Test { member: number; }`,
			Options: []interface{}{map[string]interface{}{
				"propertyDeclaration": true,
			}},
		},
		{
			Code: `interface Test { member; }`,
			Options: []interface{}{map[string]interface{}{
				"propertyDeclaration": false,
			}},
		},
		{
			Code: `type Test = { member: number; };`,
			Options: []interface{}{map[string]interface{}{
				"propertyDeclaration": true,
			}},
		},

		// Variable declarations
		{
			Code: `const x: string = '';`,
			Options: []interface{}{map[string]interface{}{
				"variableDeclaration": true,
			}},
		},
		{
			Code: `let x: string = '';`,
			Options: []interface{}{map[string]interface{}{
				"variableDeclaration": true,
			}},
		},
		{
			Code: `const a = 1;`,
			Options: []interface{}{map[string]interface{}{
				"variableDeclaration": false,
			}},
		},
		{
			Code: `const a = (): void => {};`,
			Options: []interface{}{map[string]interface{}{
				"variableDeclaration": true,
				"variableDeclarationIgnoreFunction": true,
			}},
		},
		{
			Code: `const a = function(): void {};`,
			Options: []interface{}{map[string]interface{}{
				"variableDeclaration": true,
				"variableDeclarationIgnoreFunction": true,
			}},
		},

		// Default (no options) - should not report anything
		{Code: `const [a] = [1];`},
		{Code: `const { a } = { a: 1 };`},
		{Code: `const a = 1;`},
		{Code: `class Test { state = 1; }`},
		{Code: `((a): void => {})();`},
		{Code: `function foo(a) {}`},
	}, []rule_tester.InvalidTestCase{
		// Array destructuring
		{
			Code: `const [a] = [1];`,
			Options: []interface{}{map[string]interface{}{
				"arrayDestructuring": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedef"},
			},
		},
		{
			Code: `const [a, b] = [1, 2];`,
			Options: []interface{}{map[string]interface{}{
				"arrayDestructuring": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedef"},
			},
		},

		// Object destructuring
		{
			Code: `const { a } = { a: 1 };`,
			Options: []interface{}{map[string]interface{}{
				"objectDestructuring": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedef"},
			},
		},
		{
			Code: `const { a, b } = { a: 1, b: 2 };`,
			Options: []interface{}{map[string]interface{}{
				"objectDestructuring": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedef"},
			},
		},

		// Arrow parameters
		{
			Code: `const receivesNumber = (a): void => {};`,
			Options: []interface{}{map[string]interface{}{
				"arrowParameter": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
			},
		},
		{
			Code: `const receivesStrings = (a, b): void => {};`,
			Options: []interface{}{map[string]interface{}{
				"arrowParameter": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
				{MessageId: "expectedTypedefNamed"},
			},
		},

		// Function parameters
		{
			Code: `function receivesNumber(a): void {}`,
			Options: []interface{}{map[string]interface{}{
				"parameter": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
			},
		},
		{
			Code: `function receivesStrings(a, b): void {}`,
			Options: []interface{}{map[string]interface{}{
				"parameter": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
				{MessageId: "expectedTypedefNamed"},
			},
		},
		{
			Code: `function receivesNumber([a]): void {}`,
			Options: []interface{}{map[string]interface{}{
				"parameter": true,
				"arrayDestructuring": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedef"},
			},
		},
		{
			Code: `function receivesNumbers([a, b]): void {}`,
			Options: []interface{}{map[string]interface{}{
				"parameter": true,
				"arrayDestructuring": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedef"},
			},
		},
		{
			Code: `function receivesString({ a }): void {}`,
			Options: []interface{}{map[string]interface{}{
				"parameter": true,
				"objectDestructuring": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedef"},
			},
		},
		{
			Code: `function receivesStrings({ a, b }): void {}`,
			Options: []interface{}{map[string]interface{}{
				"parameter": true,
				"objectDestructuring": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedef"},
			},
		},

		// Member variable declarations
		{
			Code: `class Test { state = 1; }`,
			Options: []interface{}{map[string]interface{}{
				"memberVariableDeclaration": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
			},
		},
		{
			Code: `class Test { state; }`,
			Options: []interface{}{map[string]interface{}{
				"memberVariableDeclaration": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
			},
		},

		// Property declarations
		{
			Code: `interface Test { member; }`,
			Options: []interface{}{map[string]interface{}{
				"propertyDeclaration": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
			},
		},
		{
			Code: `type Test = { member; };`,
			Options: []interface{}{map[string]interface{}{
				"propertyDeclaration": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
			},
		},

		// Variable declarations
		{
			Code: `const a = 1;`,
			Options: []interface{}{map[string]interface{}{
				"variableDeclaration": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
			},
		},
		{
			Code: `let a = 1;`,
			Options: []interface{}{map[string]interface{}{
				"variableDeclaration": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
			},
		},
		{
			Code: `const a = (): void => {};`,
			Options: []interface{}{map[string]interface{}{
				"variableDeclaration": true,
				"variableDeclarationIgnoreFunction": false,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
			},
		},

		// Constructor parameters
		{
			Code: `class Test { constructor(param) {} }`,
			Options: []interface{}{map[string]interface{}{
				"parameter": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
			},
		},

		// Method parameters
		{
			Code: `class Test { public method(x) { return x; } }`,
			Options: []interface{}{map[string]interface{}{
				"parameter": true,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedTypedefNamed"},
			},
		},
	})
}
