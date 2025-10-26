package prefer_destructuring

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferDestructuringRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		// Valid cases (default options)
		[]rule_tester.ValidTestCase{
			// Already using destructuring
			{Code: `const {x} = obj;`},
			{Code: `const [x] = arr;`},
			// Different names (renaming)
			{Code: `const y = obj.x;`},
			// Not a simple property access
			{Code: `const x = obj.x.y;`},
			{Code: `const x = obj[key];`},
			// Variable with type annotation (default: not enforced)
			{Code: `const x: string = obj.x;`},
			{Code: `const x: number = arr[0];`},
			// Assignment expressions (not enabled by default)
			{Code: `x = obj.x;`},
			{Code: `x = arr[0];`},
			// Complex scenarios
			{Code: `const x = getObj().prop;`},
			{Code: `const x = 1;`},
			{Code: `const x = "string";`},
		},
		// Invalid cases (default options)
		[]rule_tester.InvalidTestCase{
			// Object destructuring
			{
				Code: `const x = obj.x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
			},
			// Array destructuring
			{
				Code: `const x = arr[0];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
			},
			// Multiple variables
			{
				Code: `const x = obj.x, y = obj.y;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
					{MessageId: "preferDestructuring"},
				},
			},
		},
	)
}

func TestPreferDestructuringRuleWithTypeAnnotation(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Already using destructuring
			{Code: `const {x}: {x: string} = obj;`, Options: []interface{}{map[string]interface{}{"object": true}, map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true}}},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// With type annotation and enforce option
			{
				Code:    `const x: string = obj.x;`,
				Options: []interface{}{map[string]interface{}{"object": true}, map[string]interface{}{"enforceForDeclarationWithTypeAnnotation": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
			},
		},
	)
}

func TestPreferDestructuringRuleAssignmentExpression(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Already using destructuring
			{Code: `({x} = obj);`, Options: map[string]interface{}{"AssignmentExpression": map[string]interface{}{"object": true}}},
			{Code: `[x] = arr;`, Options: map[string]interface{}{"AssignmentExpression": map[string]interface{}{"array": true}}},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Object assignment
			{
				Code:    `x = obj.x;`,
				Options: map[string]interface{}{"AssignmentExpression": map[string]interface{}{"object": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
			},
			// Array assignment
			{
				Code:    `x = arr[0];`,
				Options: map[string]interface{}{"AssignmentExpression": map[string]interface{}{"array": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
			},
		},
	)
}

func TestPreferDestructuringRuleDisabledArrays(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		// Valid cases (arrays disabled)
		[]rule_tester.ValidTestCase{
			{Code: `const x = arr[0];`, Options: map[string]interface{}{"array": false}},
			{Code: `const x = arr[1];`, Options: map[string]interface{}{"VariableDeclarator": map[string]interface{}{"array": false}}},
		},
		// Invalid cases (objects still enabled)
		[]rule_tester.InvalidTestCase{
			{
				Code:    `const x = obj.x;`,
				Options: map[string]interface{}{"array": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
			},
		},
	)
}

func TestPreferDestructuringRuleDisabledObjects(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		// Valid cases (objects disabled)
		[]rule_tester.ValidTestCase{
			{Code: `const x = obj.x;`, Options: map[string]interface{}{"object": false}},
			{Code: `const x = obj.y;`, Options: map[string]interface{}{"VariableDeclarator": map[string]interface{}{"object": false}}},
		},
		// Invalid cases (arrays still enabled)
		[]rule_tester.InvalidTestCase{
			{
				Code:    `const x = arr[0];`,
				Options: map[string]interface{}{"object": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferDestructuring"},
				},
			},
		},
	)
}
