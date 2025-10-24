package consistent_generic_constructors

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentGenericConstructors(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentGenericConstructorsRule,
		[]rule_tester.ValidTestCase{
			// Default: prefer constructor style
			{Code: `const map = new Map<string, number>();`},
			{Code: `const set = new Set<string>();`},

			// Both sides have type args (allowed)
			{Code: `const map: Map<string, number> = new Map<string, number>();`},

			// Neither side has type args (allowed)
			{Code: `const map: Map = new Map();`},
			{Code: `const map = new Map();`},

			// With type-annotation preference
			{
				Code:    `const map: Map<string, number> = new Map();`,
				Options: "type-annotation",
			},

			// Non-generic constructors
			{Code: `const date = new Date();`},
			{Code: `const obj = new Object();`},
		},
		[]rule_tester.InvalidTestCase{
			// Default (constructor): type args on annotation only
			{
				Code:   `const map: Map<string, number> = new Map();`,
				Output: []string{`const map: Map = new Map<string, number>();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferConstructor"},
				},
			},

			// Constructor preference: with Set
			{
				Code:   `const set: Set<string> = new Set();`,
				Output: []string{`const set: Set = new Set<string>();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferConstructor"},
				},
			},

			// Type-annotation preference: type args on constructor only
			{
				Code:    `const map = new Map<string, number>();`,
				Output:  []string{`const map: Map<string, number> = new Map();`},
				Options: "type-annotation",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTypeAnnotation"},
				},
			},

			// Type-annotation preference: with Set
			{
				Code:    `const set = new Set<string>();`,
				Output:  []string{`const set: Set<string> = new Set();`},
				Options: "type-annotation",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferTypeAnnotation"},
				},
			},
		},
	)
}
