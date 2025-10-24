package consistent_type_definitions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTypeDefinitions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		[]rule_tester.ValidTestCase{
			// Default: prefer interface
			{Code: `interface T { x: number }`},
			{Code: `interface T { x: number; y: string }`},

			// Type preference
			{
				Code:    `type T = { x: number }`,
				Options: "type",
			},
			{
				Code:    `type T = { x: number; y: string }`,
				Options: "type",
			},

			// Non-object types are allowed regardless of preference
			{Code: `type T = string;`},
			{Code: `type T = string | number;`},
			{Code: `type T = Array<string>;`},
		},
		[]rule_tester.InvalidTestCase{
			// Default (interface): type used for object
			{
				Code: `type T = { x: number }`,
				Errors: []rule_tester.ExpectedError{
					{MessageId: "interfaceOverType"},
				},
			},

			// Default (interface): multiple properties
			{
				Code: `type T = { x: number; y: string }`,
				Errors: []rule_tester.ExpectedError{
					{MessageId: "interfaceOverType"},
				},
			},

			// Type preference: interface used
			{
				Code:    `interface T { x: number }`,
				Options: "type",
				Errors: []rule_tester.ExpectedError{
					{MessageId: "typeOverInterface"},
				},
			},

			// Type preference: interface with multiple properties
			{
				Code:    `interface T { x: number; y: string }`,
				Options: "type",
				Errors: []rule_tester.ExpectedError{
					{MessageId: "typeOverInterface"},
				},
			},
		},
		t,
		&ConsistentTypeDefinitionsRule,
	)
}
