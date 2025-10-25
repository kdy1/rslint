package consistent_type_definitions

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestConsistentTypeDefinitionsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentTypeDefinitionsRule,
		[]rule_tester.ValidTestCase{
			// Valid with default option (prefer: "interface")
			{Code: `interface T { x: number; }`},
			{Code: `type T = string;`}, // Non-object types are allowed
			{Code: `type T = string | number;`},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid with default option (prefer: "interface")
			{
				Code: `type T = { x: number; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "typeOverInterface",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{`interface T { x: number; }`},
			},
		},
	)
}

func TestConsistentTypeDefinitionsRuleWithOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentTypeDefinitionsRule,
		[]rule_tester.ValidTestCase{
			{
				Code: `type T = { x: number; };`,
				Options: map[string]interface{}{
					"prefer": "type",
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `interface T { x: number; }`,
				Options: map[string]interface{}{
					"prefer": "type",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "interfaceOverType",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{`type T = { x: number; }`},
			},
		},
	)
}
