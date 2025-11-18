package no_restricted_imports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRestrictedImportsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedImportsRule, []rule_tester.ValidTestCase{
		// Valid imports without restrictions
		{Code: "import foo from 'foo';"},
		{Code: "import foo = require('foo');"},
		{Code: "import 'foo';"},
		{
			Code:    "import foo from 'foo';",
			Options: []interface{}{"import1", "import2"},
		},
		{
			Code:    "import foo = require('foo');",
			Options: []interface{}{"import1", "import2"},
		},
		{
			Code:    "export { foo } from 'foo';",
			Options: []interface{}{"import1", "import2"},
		},
		{
			Code:    "import foo from 'foo';",
			Options: []interface{}{map[string]interface{}{"paths": []interface{}{"import1", "import2"}}},
		},
		{
			Code:    "export { foo } from 'foo';",
			Options: []interface{}{map[string]interface{}{"paths": []interface{}{"import1", "import2"}}},
		},
		{
			Code:    "import 'foo';",
			Options: []interface{}{"import1", "import2"},
		},

		// Patterns with negations
		{
			Code: "import foo from 'foo';",
			Options: []interface{}{map[string]interface{}{
				"paths":    []interface{}{"import1", "import2"},
				"patterns": []interface{}{"import1/private/*", "import2/*", "!import2/good"},
			}},
		},
		{
			Code: "export { foo } from 'foo';",
			Options: []interface{}{map[string]interface{}{
				"paths":    []interface{}{"import1", "import2"},
				"patterns": []interface{}{"import1/private/*", "import2/*", "!import2/good"},
			}},
		},

		// Custom messages
		{
			Code: "import foo from 'foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":    "import-foo",
						"message": "Please use import-bar instead.",
					},
					map[string]interface{}{
						"name":    "import-baz",
						"message": "Please use import-quux instead.",
					},
				},
			}},
		},
		{
			Code: "export { foo } from 'foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":    "import-foo",
						"message": "Please use import-bar instead.",
					},
					map[string]interface{}{
						"name":    "import-baz",
						"message": "Please use import-quux instead.",
					},
				},
			}},
		},

		// Import names restrictions (not matching)
		{
			Code: "import foo from 'foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "import-foo",
						"importNames": []interface{}{"Bar"},
						"message":     "Please use Bar from /import-bar/baz/ instead.",
					},
				},
			}},
		},
		{
			Code: "export { foo } from 'foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "import-foo",
						"importNames": []interface{}{"Bar"},
						"message":     "Please use Bar from /import-bar/baz/ instead.",
					},
				},
			}},
		},

		// Pattern groups
		{
			Code: "import foo from 'foo';",
			Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":   []interface{}{"import1/private/*"},
						"message": "usage of import1 private modules not allowed.",
					},
					map[string]interface{}{
						"group":   []interface{}{"import2/*", "!import2/good"},
						"message": "import2 is deprecated, except the modules in import2/good.",
					},
				},
			}},
		},
		{
			Code: "export { foo } from 'foo';",
			Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":   []interface{}{"import1/private/*"},
						"message": "usage of import1 private modules not allowed.",
					},
					map[string]interface{}{
						"group":   []interface{}{"import2/*", "!import2/good"},
						"message": "import2 is deprecated, except the modules in import2/good.",
					},
				},
			}},
		},

		// Type imports with allowTypeImports
		{
			Code: "import type foo from 'import-foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"message":          "Please use import-bar instead.",
						"allowTypeImports": true,
					},
				},
			}},
		},
		{
			Code: "import type _ = require('import-foo');",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"message":          "Please use import-bar instead.",
						"allowTypeImports": true,
					},
				},
			}},
		},
		{
			Code: "import type { Bar } from 'import-foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"message":          "Please use import-bar instead.",
						"allowTypeImports": true,
					},
				},
			}},
		},
		{
			Code: "export type { Bar } from 'import-foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"message":          "Please use import-bar instead.",
						"allowTypeImports": true,
					},
				},
			}},
		},

		// Import from good path within restricted pattern
		{
			Code: "import foo from 'import2/good';",
			Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":   []interface{}{"import2/*", "!import2/good"},
						"message": "import2 is deprecated, except the modules in import2/good.",
					},
				},
			}},
		},
	}, []rule_tester.InvalidTestCase{
		// Simple string array restrictions
		{
			Code:    "import foo from 'import1';",
			Options: []interface{}{"import1", "import2"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    "import 'import1';",
			Options: []interface{}{"import1", "import2"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    "export * from 'import1';",
			Options: []interface{}{"import1", "import2"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Path restrictions with custom messages
		{
			Code: "import foo from 'import-foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":    "import-foo",
						"message": "Please use import-bar instead.",
					},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "pathWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "export { foo } from 'import-foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":    "import-foo",
						"message": "Please use import-bar instead.",
					},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "pathWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Import names restrictions
		{
			Code: "import { Bar } from 'import-foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "import-foo",
						"importNames": []interface{}{"Bar"},
						"message":     "Please use Bar from /import-bar/baz/ instead.",
					},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "pathWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "export { Bar } from 'import-foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "import-foo",
						"importNames": []interface{}{"Bar"},
						"message":     "Please use Bar from /import-bar/baz/ instead.",
					},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "pathWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Pattern restrictions
		{
			Code: "import foo from 'import1/private/foo';",
			Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":   []interface{}{"import1/private/*"},
						"message": "usage of import1 private modules not allowed.",
					},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "export * from 'import1/private/foo';",
			Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":   []interface{}{"import1/private/*"},
						"message": "usage of import1 private modules not allowed.",
					},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Pattern with negation (should error for non-good paths)
		{
			Code: "import foo from 'import2/bad';",
			Options: []interface{}{map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":   []interface{}{"import2/*", "!import2/good"},
						"message": "import2 is deprecated, except the modules in import2/good.",
					},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Type imports without allowTypeImports
		{
			Code: "import type foo from 'import-foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":    "import-foo",
						"message": "Please use import-bar instead.",
					},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "pathWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Regular imports with allowTypeImports (should still error)
		{
			Code: "import foo from 'import-foo';",
			Options: []interface{}{map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"message":          "Please use import-bar instead.",
						"allowTypeImports": true,
					},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "pathWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Import equals (import = require)
		{
			Code:    "import foo = require('import1');",
			Options: []interface{}{"import1", "import2"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Multiple errors
		{
			Code: `
import foo from 'import1';
import bar from 'import2';
			`,
			Options: []interface{}{"import1", "import2"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      2,
					Column:    1,
				},
				{
					MessageId: "path",
					Line:      3,
					Column:    1,
				},
			},
		},
	})
}
