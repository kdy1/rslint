package no_import_type_side_effects

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoImportTypeSideEffectsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoImportTypeSideEffectsRule,
		[]rule_tester.ValidTestCase{
			// Valid: regular value import
			{Code: "import T from 'mod';"},

			// Valid: namespace import
			{Code: "import * as T from 'mod';"},

			// Valid: named value import
			{Code: "import { T } from 'mod';"},

			// Valid: already using top-level type import
			{Code: "import type { T } from 'mod';"},

			// Valid: top-level type import with multiple specifiers
			{Code: "import type { T, U } from 'mod';"},

			// Valid: mixed type and value imports (inline type is necessary)
			{Code: "import { type T, U } from 'mod';"},

			// Valid: mixed value and type imports
			{Code: "import { T, type U } from 'mod';"},

			// Valid: top-level type import for default
			{Code: "import type T from 'mod';"},

			// Valid: mixed default value and inline type
			{Code: "import T, { type U } from 'mod';"},

			// Valid: top-level type with default and named
			{Code: "import type T, { U } from 'mod';"},

			// Valid: namespace type import
			{Code: "import type * as T from 'mod';"},

			// Valid: side-effect only import
			{Code: "import 'mod';"},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid: single inline type import
			{
				Code: "import { type A } from 'mod';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useTopLevelQualifier"},
				},
				Output: []string{"import type { A } from 'mod';"},
			},

			// Invalid: single inline type import with alias
			{
				Code: "import { type A as AA } from 'mod';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useTopLevelQualifier"},
				},
				Output: []string{"import type { A as AA } from 'mod';"},
			},

			// Invalid: multiple inline type imports
			{
				Code: "import { type A, type B } from 'mod';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useTopLevelQualifier"},
				},
				Output: []string{"import type { A, B } from 'mod';"},
			},

			// Invalid: multiple inline type imports with aliases
			{
				Code: "import { type A as AA, type B as BB } from 'mod';",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useTopLevelQualifier"},
				},
				Output: []string{"import type { A as AA, B as BB } from 'mod';"},
			},
		},
	)
}
