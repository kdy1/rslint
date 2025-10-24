package consistent_indexed_object_style

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentIndexedObjectStyle(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		[]rule_tester.ValidTestCase{
			// Default: prefer Record
			{Code: `type T = Record<string, unknown>;`},
			{Code: `type T = Record<string, number>;`},

			// Index signature with index-signature preference
			{
				Code:    `type T = { [key: string]: unknown };`,
				Options: "index-signature",
			},
			{
				Code:    `interface I { [key: string]: number }`,
				Options: "index-signature",
			},

			// Interfaces with additional members (not pure index signature)
			{Code: `interface I { [key: string]: number; x: number }`},

			// Regular types
			{Code: `type T = { x: number };`},
			{Code: `interface I { x: number }`},
		},
		[]rule_tester.InvalidTestCase{
			// Default (record): index signature in type
			{
				Code: `type T = { [key: string]: unknown };`,
				Errors: []rule_tester.ExpectedError{
					{MessageId: "preferRecord"},
				},
			},

			// Default (record): index signature in interface
			{
				Code: `interface I { [key: string]: number }`,
				Errors: []rule_tester.ExpectedError{
					{MessageId: "preferRecord"},
				},
			},

			// Index-signature preference: Record type
			{
				Code:    `type T = Record<string, unknown>;`,
				Options: "index-signature",
				Errors: []rule_tester.ExpectedError{
					{MessageId: "preferIndexSignature"},
				},
			},

			// With different key types
			{
				Code: `type T = { [key: number]: string };`,
				Errors: []rule_tester.ExpectedError{
					{MessageId: "preferRecord"},
				},
			},
		},
		t,
		&ConsistentIndexedObjectStyleRule,
	)
}
