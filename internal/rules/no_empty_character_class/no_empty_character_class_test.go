package no_empty_character_class

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoEmptyCharacterClassRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyCharacterClassRule,
		[]rule_tester.ValidTestCase{
			// Valid patterns - character class with range
			{Code: `const regex = /^abc[a-zA-Z]/`},

			// RegExp constructor
			{Code: `const regex = new RegExp("^abc[]")`},

			// No character class
			{Code: `const regex = /^abc/`},

			// Escaped brackets
			{Code: `const regex = /[\\[]/`},
			{Code: `const regex = /[\\]]/`},
			{Code: `const regex = /\\[][\\]]/`},
			{Code: `const regex = /[a-zA-Z\\[]/`},
			{Code: `const regex = /[[]/`},
			{Code: `const regex = /[\\[a-z[]]/`},
			{Code: `const regex = /[\\-\\[\\]\\/\\{\\}\\(\\)\\*\\+\\?\\.\\\\^\\$\\|]/g`},

			// Whitespace patterns
			{Code: `const regex = /\\s*:\\s*/gim`},

			// Negated empty class (allowed)
			{Code: `const regex = /[^]/`},
			{Code: `const regex = /\\[][^]/`},

			// ES6+ flags
			{Code: `const regex = /[\\]]/uy`},
			{Code: `const regex = /[\\]]/s`},
			{Code: `const regex = /[\\]]/d`},

			// ES2024 v-flag patterns
			{Code: `const regex = /[[^]]/v`},
			{Code: `const regex = /[[\\]]]/v`},
			{Code: `const regex = /[[\\[]]/v`},
			{Code: `const regex = /[a--b]/v`},
			{Code: `const regex = /[a&&b]/v`},
			{Code: `const regex = /[[a][b]]/v`},
			{Code: `const regex = /[\\q{}]/v`},
			{Code: `const regex = /[[^]--\\p{ASCII}]/v`},
		},
		[]rule_tester.InvalidTestCase{
			// Empty character class
			{
				Code: `const regex = /^abc[]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Empty class between text
			{
				Code: `const regex = /foo[]bar/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Empty class with trailing bracket
			{
				Code: `const regex = /[]]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Escaped bracket then empty class
			{
				Code: `const regex = /\\[[]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Empty class at pattern end
			{
				Code: `const regex = /\\[\\[\\]a-z[]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// ES2022 with empty class
			{
				Code: `const regex = /[]]/d`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// ES2015 unicode with empty class
			{
				Code: `const regex = /[(]\\u{0}*[]/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// ES2024 v-flag patterns with empty classes
			{
				Code: `const regex = /[]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `const regex = /[[]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `const regex = /[[a][]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `const regex = /[a[[b[]c]]d]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `const regex = /[a--[]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `const regex = /[[]--b]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `const regex = /[a&&[]]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `const regex = /[[]&&b]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
		},
	)
}
