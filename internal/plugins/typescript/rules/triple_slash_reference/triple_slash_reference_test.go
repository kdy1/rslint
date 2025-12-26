package triple_slash_reference

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestTripleSlashReferenceRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&TripleSlashReferenceRule,
		[]rule_tester.ValidTestCase{
			// Double-slash comments with 'never' options - Case 1
			{
				Code: `
// <reference path="foo" />
// <reference types="bar" />
// <reference lib="baz" />
import * as foo from 'foo';
import * as bar from 'bar';
import * as baz from 'baz';
`,
				Options: map[string]interface{}{
					"lib":   "never",
					"path":  "never",
					"types": "never",
				},
			},
			// Double-slash comments with CommonJS - Case 2
			{
				Code: `
// <reference path="foo" />
// <reference types="bar" />
// <reference lib="baz" />
import foo = require('foo');
import bar = require('bar');
import baz = require('baz');
`,
				Options: map[string]interface{}{
					"lib":   "never",
					"path":  "never",
					"types": "never",
				},
			},
			// Triple-slash comments with ES6 imports and 'always' options - Case 3
			{
				Code: `
/// <reference path="foo" />
/// <reference types="bar" />
/// <reference lib="baz" />
import * as foo from 'foo';
import * as bar from 'bar';
import * as baz from 'baz';
`,
				Options: map[string]interface{}{
					"lib":   "always",
					"path":  "always",
					"types": "always",
				},
			},
			// Triple-slash comments with CommonJS and 'always' options - Case 4
			{
				Code: `
/// <reference path="foo" />
/// <reference types="bar" />
/// <reference lib="baz" />
import foo = require('foo');
import bar = require('bar');
import baz = require('baz');
`,
				Options: map[string]interface{}{
					"lib":   "always",
					"path":  "always",
					"types": "always",
				},
			},
			// Triple-slash with namespace imports (simple) - Case 5
			{
				Code: `
/// <reference path="foo" />
/// <reference types="bar" />
/// <reference lib="baz" />
import foo = foo;
import bar = bar;
import baz = baz;
`,
				Options: map[string]interface{}{
					"lib":   "always",
					"path":  "always",
					"types": "always",
				},
			},
			// Triple-slash with namespace imports (nested) - Case 6
			{
				Code: `
/// <reference path="foo" />
/// <reference types="bar" />
/// <reference lib="baz" />
import foo = foo.foo;
import bar = bar.bar.bar.bar;
import baz = baz.baz;
`,
				Options: map[string]interface{}{
					"lib":   "always",
					"path":  "always",
					"types": "always",
				},
			},
			// Single option 'path' with ES6 import - Case 7
			{
				Code: "import * as foo from 'foo';",
				Options: map[string]interface{}{
					"path": "never",
				},
			},
			// Single option 'path' with CommonJS - Case 8
			{
				Code: "import foo = require('foo');",
				Options: map[string]interface{}{
					"path": "never",
				},
			},
			// Single option 'types' with ES6 import - Case 9
			{
				Code: "import * as foo from 'foo';",
				Options: map[string]interface{}{
					"types": "never",
				},
			},
			// Single option 'types' with CommonJS - Case 10
			{
				Code: "import foo = require('foo');",
				Options: map[string]interface{}{
					"types": "never",
				},
			},
			// Single option 'lib' with ES6 import - Case 11
			{
				Code: "import * as foo from 'foo';",
				Options: map[string]interface{}{
					"lib": "never",
				},
			},
			// Single option 'lib' with CommonJS - Case 12
			{
				Code: "import foo = require('foo');",
				Options: map[string]interface{}{
					"lib": "never",
				},
			},
			// Prefer-import with ES6 import - Case 13
			{
				Code: "import * as foo from 'foo';",
				Options: map[string]interface{}{
					"types": "prefer-import",
				},
			},
			// Prefer-import with CommonJS - Case 14
			{
				Code: "import foo = require('foo');",
				Options: map[string]interface{}{
					"types": "prefer-import",
				},
			},
			// Prefer-import with different reference and import - Case 15
			{
				Code: `
/// <reference types="foo" />
import * as bar from 'bar';
`,
				Options: map[string]interface{}{
					"types": "prefer-import",
				},
			},
			// Commented-out references in block comments - Case 16
			{
				Code: `
/*
/// <reference types="foo" />
*/
import * as foo from 'foo';
`,
				Options: map[string]interface{}{
					"lib":   "never",
					"path":  "never",
					"types": "never",
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Triple-slash types with prefer-import (ES6 import) - Case 1
			{
				Code: `
/// <reference types="foo" />
import * as foo from 'foo';
`,
				Options: map[string]interface{}{
					"types": "prefer-import",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tripleSlashReference",
						Line:      2,
						Column:    1,
					},
				},
			},
			// Triple-slash types with prefer-import (CommonJS require) - Case 2
			{
				Code: `
/// <reference types="foo" />
import foo = require('foo');
`,
				Options: map[string]interface{}{
					"types": "prefer-import",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tripleSlashReference",
						Line:      2,
						Column:    1,
					},
				},
			},
			// Triple-slash path when never allowed - Case 3
			{
				Code: `/// <reference path="foo" />`,
				Options: map[string]interface{}{
					"path": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tripleSlashReference",
						Line:      1,
						Column:    1,
					},
				},
			},
			// Triple-slash types when never allowed - Case 4
			{
				Code: `/// <reference types="foo" />`,
				Options: map[string]interface{}{
					"types": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tripleSlashReference",
						Line:      1,
						Column:    1,
					},
				},
			},
			// Triple-slash lib when never allowed - Case 5
			{
				Code: `/// <reference lib="foo" />`,
				Options: map[string]interface{}{
					"lib": "never",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tripleSlashReference",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}
