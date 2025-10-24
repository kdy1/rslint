package naming_convention

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNamingConventionRule(t *testing.T) {
	// TODO: Implement comprehensive tests
	// Test cases should cover:
	// 1. Different selectors (variable, function, class, interface, etc.)
	// 2. Different formats (camelCase, PascalCase, snake_case, UPPER_CASE)
	// 3. Modifiers (private, protected, public, static, readonly, const, etc.)
	// 4. Leading/trailing underscore options
	// 5. Prefix/suffix requirements
	// 6. Custom regex patterns
	// 7. Type-specific rules (boolean, function, array, etc.)
	// 8. Filter patterns

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NamingConventionRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `const myVariable = 1;`,
				Options: []interface{}{
					map[string]interface{}{
						"selector": "variable",
						"format":   []string{"camelCase"},
					},
				},
			},
			{
				Code: `class MyClass {}`,
				Options: []interface{}{
					map[string]interface{}{
						"selector": "class",
						"format":   []string{"PascalCase"},
					},
				},
			},
			{
				Code: `interface IMyInterface {}`,
				Options: []interface{}{
					map[string]interface{}{
						"selector": "interface",
						"format":   []string{"PascalCase"},
						"prefix":   []string{"I"},
					},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases
			// These should test various violations of naming conventions
		},
	)
}
