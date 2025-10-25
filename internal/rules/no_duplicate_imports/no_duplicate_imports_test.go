package no_duplicate_imports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoDuplicateImportsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDuplicateImportsRule,
		[]rule_tester.ValidTestCase{
			// Basic imports from different modules
			{Code: `import { merge } from 'lodash-es'; import { find } from 'rxjs';`},

			// Named and namespace imports from different modules
			{Code: `import { merge } from 'lodash-es'; import * as utils from 'rxjs';`},

			// Side-effect imports from different modules
			{Code: `import 'polyfill1'; import 'polyfill2';`},

			// Default and named imports from different modules
			{Code: `import foo from 'foo'; import bar from 'bar';`},

			// Combined import patterns from different modules
			{Code: `import foo, { bar } from 'foo'; import baz from 'baz';`},

			// Exports from different modules (without includeExports)
			{Code: `export { foo } from 'foo'; export { bar } from 'bar';`},

			// Mix of imports and exports (without includeExports)
			{Code: `import { foo } from 'foo'; export { bar } from 'bar';`},
		},
		[]rule_tester.InvalidTestCase{
			// Duplicate side-effect imports
			{
				Code: `import 'fs'; import 'fs';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "import", Line: 1, Column: 24},
				},
			},

			// Multiple named imports from the same module
			{
				Code: `import { merge } from 'lodash-es'; import { find } from 'lodash-es';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "import", Line: 1, Column: 57},
				},
			},

			// Three separate imports from same module
			{
				Code: `import foo from 'os'; import { bar } from 'os'; import * as baz from 'os';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "import", Line: 1, Column: 39},
					{MessageId: "import", Line: 1, Column: 63},
				},
			},

			// Duplicate default imports
			{
				Code: `import foo from 'foo'; import bar from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "import", Line: 1, Column: 36},
				},
			},

			// Duplicate namespace imports
			{
				Code: `import * as foo from 'foo'; import * as bar from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "import", Line: 1, Column: 46},
				},
			},
		},
	)
}

func TestNoDuplicateImportsRuleWithIncludeExports(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDuplicateImportsRule,
		[]rule_tester.ValidTestCase{
			// Exports from different modules
			{
				Code:    `export { foo } from 'foo'; export { bar } from 'bar';`,
				Options: map[string]interface{}{"includeExports": true},
			},

			// Import and export from different modules
			{
				Code:    `import { foo } from 'foo'; export { bar } from 'bar';`,
				Options: map[string]interface{}{"includeExports": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Duplicate exports from same module
			{
				Code: `export { foo } from 'foo'; export { bar } from 'foo';`,
				Options: map[string]interface{}{
					"includeExports": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "export", Line: 1, Column: 48},
				},
			},

			// Import then export from same module
			{
				Code: `import { foo } from 'foo'; export { bar } from 'foo';`,
				Options: map[string]interface{}{
					"includeExports": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "exportImport", Line: 1, Column: 48},
				},
			},

			// Export then import from same module
			{
				Code: `export { foo } from 'foo'; import { bar } from 'foo';`,
				Options: map[string]interface{}{
					"includeExports": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "importExport", Line: 1, Column: 48},
				},
			},

			// Multiple duplicate exports
			{
				Code: `export { foo } from 'foo'; export { bar } from 'bar'; export { baz } from 'foo';`,
				Options: map[string]interface{}{
					"includeExports": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "export", Line: 1, Column: 75},
				},
			},
		},
	)
}
