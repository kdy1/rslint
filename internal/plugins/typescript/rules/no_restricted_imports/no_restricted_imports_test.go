package no_restricted_imports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRestrictedImportsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedImportsRule, []rule_tester.ValidTestCase{
		// No restrictions configured
		{Code: `import foo from "foo";`},
		{Code: `import { bar } from "foo";`},

		// Paths restriction - allowed imports
		{
			Code:    `import foo from "bar";`,
			Options: map[string]interface{}{"paths": []string{"foo"}},
		},
		{
			Code:    `import { baz } from "bar";`,
			Options: map[string]interface{}{"paths": []string{"foo"}},
		},

		// Patterns restriction - allowed imports
		{
			Code:    `import foo from "foo/bar";`,
			Options: map[string]interface{}{"patterns": []string{"foo"}},
		},
		{
			Code: `import foo from "foo/bar";`,
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{"group": "foo/baz"},
				},
			},
		},

		// AllowTypeImports - type-only imports allowed
		{
			Code: `import type foo from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "foo",
						"allowTypeImports": true,
					},
				},
			},
		},
		{
			Code: `import type { bar } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "foo",
						"allowTypeImports": true,
					},
				},
			},
		},
		{
			Code: `import { type bar } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "foo",
						"importNames":      []string{"bar"},
						"allowTypeImports": true,
					},
				},
			},
		},

		// ImportNames - only specific names restricted
		{
			Code: `import { baz } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "foo",
						"importNames": []string{"bar"},
					},
				},
			},
		},
		{
			Code: `import foo from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "foo",
						"importNames": []string{"bar"},
					},
				},
			},
		},

		// Multiple paths
		{
			Code:    `import foo from "baz";`,
			Options: map[string]interface{}{"paths": []string{"foo", "bar"}},
		},

		// Case sensitivity in patterns
		{
			Code: `import foo from "FOO";`,
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":         "foo",
						"caseSensitive": true,
					},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Basic path restriction
		{
			Code:    `import foo from "foo";`,
			Options: map[string]interface{}{"paths": []string{"foo"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `import { bar } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{"name": "foo"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Path restriction with custom message
		{
			Code: `import foo from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":    "foo",
						"message": "Please use 'bar' instead of 'foo'.",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "pathWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Pattern restriction
		{
			Code:    `import foo from "foo/bar";`,
			Options: map[string]interface{}{"patterns": []string{"foo/*"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `import foo from "foo/bar";`,
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{"group": "foo/*"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Pattern restriction with custom message
		{
			Code: `import foo from "foo/bar";`,
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":   "foo/*",
						"message": "Do not import from foo subdirectories.",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patternWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},

		// ImportNames restriction
		{
			Code: `import { bar } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "foo",
						"importNames": []string{"bar"},
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importName",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `import { bar } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "foo",
						"importNames": []string{"bar"},
						"message":     "bar is deprecated, use baz instead.",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importNameWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},

		// AllowTypeImports - value imports still restricted
		{
			Code: `import foo from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "foo",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `import { bar } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "foo",
						"importNames":      []string{"bar"},
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importName",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Export declarations
		{
			Code:    `export { foo } from "foo";`,
			Options: map[string]interface{}{"paths": []string{"foo"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `export * from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{"name": "foo"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Pattern with importNames
		{
			Code: `import { bar } from "foo/baz";`,
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":       "foo/*",
						"importNames": []string{"bar"},
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importName",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Case insensitive patterns
		{
			Code: `import foo from "FOO";`,
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":         "foo",
						"caseSensitive": false,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Wildcard patterns
		{
			Code:    `import foo from "foo/bar/baz";`,
			Options: map[string]interface{}{"patterns": []string{"foo/**"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Multiple import specifiers with one restricted
		{
			Code: `import { bar, baz } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "foo",
						"importNames": []string{"bar"},
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importName",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Import with alias
		{
			Code: `import { bar as myBar } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "foo",
						"importNames": []string{"bar"},
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importName",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Namespace import
		{
			Code:    `import * as foo from "foo";`,
			Options: map[string]interface{}{"paths": []string{"foo"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Default and named imports
		{
			Code: `import foo, { bar } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "foo",
						"importNames": []string{"bar"},
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importName",
					Line:      1,
					Column:    1,
				},
			},
		},
	})
}

func TestNoRestrictedImportsRuleTypeOnlyExports(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedImportsRule, []rule_tester.ValidTestCase{
		// Type-only export allowed
		{
			Code: `export type { foo } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "foo",
						"allowTypeImports": true,
					},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Value export restricted
		{
			Code: `export { foo } from "foo";`,
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "foo",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
	})
}

func TestNoRestrictedImportsRuleMultipleRestrictions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedImportsRule, []rule_tester.ValidTestCase{
		{
			Code: `import baz from "baz";`,
			Options: map[string]interface{}{
				"paths":    []string{"foo", "bar"},
				"patterns": []string{"lodash/*"},
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `import foo from "foo";`,
			Options: map[string]interface{}{
				"paths":    []string{"foo", "bar"},
				"patterns": []string{"lodash/*"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `import bar from "bar";`,
			Options: map[string]interface{}{
				"paths":    []string{"foo", "bar"},
				"patterns": []string{"lodash/*"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `import map from "lodash/map";`,
			Options: map[string]interface{}{
				"paths":    []string{"foo", "bar"},
				"patterns": []string{"lodash/*"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},
	})
}

func TestNoRestrictedImportsRuleComplexPatterns(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedImportsRule, []rule_tester.ValidTestCase{
		// Allowed patterns
		{
			Code:    `import foo from "foo";`,
			Options: map[string]interface{}{"patterns": []string{"foo/*"}},
		},
		{
			Code:    `import foo from "foo/bar/baz";`,
			Options: map[string]interface{}{"patterns": []string{"bar/*"}},
		},
	}, []rule_tester.InvalidTestCase{
		// Pattern matching
		{
			Code:    `import foo from "foo/bar";`,
			Options: map[string]interface{}{"patterns": []string{"foo/*"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    `import foo from "foo/bar/baz";`,
			Options: map[string]interface{}{"patterns": []string{"foo/**"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},
	})
}
