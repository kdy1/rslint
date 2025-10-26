package no_invalid_regexp

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidRegexpRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidRegexpRule,
		[]rule_tester.ValidTestCase{
			// Basic valid patterns
			{Code: `RegExp('');`},
			{Code: `RegExp();`},
			{Code: `RegExp('.', 'g');`},
			{Code: `new RegExp('.');`},
			{Code: `new RegExp();`},
			{Code: `new RegExp('.', 'im');`},
			{Code: `new RegExp('.', 'y');`},
			{Code: `new RegExp('.', 'u');`},
			{Code: `new RegExp('.', 'yu');`},
			{Code: `new RegExp('/', 'yu');`},
			{Code: `new RegExp('\/', 'yu');`},
			{Code: `new RegExp('.', 'y');`},
			{Code: `new RegExp('.', 'u');`},
			{Code: `new RegExp('.', 'yu');`},
			{Code: `new RegExp('/', 'yu');`},
			{Code: `new RegExp('\/', 'yu');`},
			{Code: `new RegExp('.', 's');`},
			{Code: `new RegExp('(?<a>b)\\k<a>');`},
			{Code: `new RegExp('(?<a>b)\\k<a>', 'u');`},
			{Code: `new RegExp('(?<=a)b');`},
			{Code: `new RegExp('(?<!a)b');`},
			{Code: `new RegExp('(?<a>b)\\k<a>', 'u');`},
			{Code: `new RegExp('(?<=a)b');`},
			{Code: `new RegExp('(?<!a)b');`},
			{Code: `new RegExp('\\\\u{0}');`},
			{Code: `new RegExp('\\\\u{0}*');`},
			{Code: `new RegExp('\\\\u{0}*', 'u');`},
			{Code: `new RegExp('[\\\\u{0}-\\\\u{1F}]', 'u');`},
			{Code: `new RegExp('.', 'd');`},
			{Code: `new RegExp('\\\\p{Letter}', 'u');`},

			// ES2024 v-flag tests
			{Code: `new RegExp('[a--b]', 'v');`},
			{Code: `new RegExp('[a&&b]', 'v');`},
			{Code: `new RegExp('[a--[0-9]]', 'v');`},
			{Code: `new RegExp('[\\\\p{Basic_Emoji}--\\\\q{a|bc|def}]', 'v');`},
			{Code: `new RegExp('[a--b]', 'v');`},
			{Code: `new RegExp('[a&&b]', 'v');`},
			{Code: `new RegExp('[a--[0-9]]', 'v');`},
			{Code: `new RegExp('[\\\\p{Basic_Emoji}--\\\\q{a|bc|def}]', 'v');`},

			// Unicode escapes
			{Code: `new RegExp('\\\\u{65}', 'u');`},
			{Code: `new RegExp('\\\\u{65}*', 'u');`},
			{Code: `new RegExp('[\\\\u{0}-\\\\u{1F}]', 'u');`},

			// Dynamic flags or patterns (not validated)
			{Code: `new RegExp('.', flags);`},
			{Code: `new RegExp(pattern, 'g');`},
			{Code: `new RegExp(pattern);`},

			// Named groups
			{Code: `new RegExp('(?<a>b)');`},
			{Code: `new RegExp('(?<a>b)\\\\k<a>');`},

			// Valid flag combinations
			{Code: `new RegExp('.', 'gi');`},
			{Code: `new RegExp('.', 'gim');`},
			{Code: `new RegExp('.', 'gimsuy');`},
			{Code: `new RegExp('.', 'ag');`},
			{Code: `new RegExp('.', 'ga');`},
			{Code: `new RegExp('.', 'az');`},
			{Code: `new RegExp('.', 'za');`},
			{Code: `new RegExp('.', 'agz');`},
		},
		[]rule_tester.InvalidTestCase{
			// Unterminated character class
			{
				Code: `RegExp('[');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Unmatched closing paren
			{
				Code: `new RegExp(')');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Trailing backslash
			{
				Code: `new RegExp('\\\\');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Invalid flag 'z'
			{
				Code: `RegExp('.', 'z');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Invalid flag combination 'uv'
			{
				Code: `new RegExp('.', 'uv');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Duplicate flags
			{
				Code: `RegExp('.', 'aa');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			{
				Code: `new RegExp('.', 'uu');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			{
				Code: `new RegExp('.', 'aaz');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			{
				Code: `new RegExp('.', 'aga');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Invalid escape in unicode mode
			{
				Code: `new RegExp('\\\\a', 'u');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Nothing to repeat
			{
				Code: `RegExp('*');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			{
				Code: `new RegExp('+');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Invalid quantifier
			{
				Code: `new RegExp('?');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Unmatched opening paren
			{
				Code: `new RegExp('(');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Invalid case-sensitive duplicate flags
			{
				Code: `new RegExp('.', 'aA');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Multiple errors in one test - should report first error
			{
				Code: `new RegExp('[', 'gg');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
		},
	)
}

func TestNoInvalidRegexpRuleWithOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidRegexpRule,
		[]rule_tester.ValidTestCase{
			// Allow custom constructor flag 'a'
			{
				Code: `RegExp('.', 'a');`,
				Options: map[string]interface{}{
					"allowConstructorFlags": []interface{}{"a"},
				},
			},
			// Allow custom constructor flags 'a' and 'z'
			{
				Code: `RegExp('.', 'z');`,
				Options: map[string]interface{}{
					"allowConstructorFlags": []interface{}{"a", "z"},
				},
			},
			// Allow combination of custom and standard flags
			{
				Code: `new RegExp('.', 'agz');`,
				Options: map[string]interface{}{
					"allowConstructorFlags": []interface{}{"a", "z"},
				},
			},
			// Empty allowConstructorFlags
			{
				Code: `new RegExp('.', 'g');`,
				Options: map[string]interface{}{
					"allowConstructorFlags": []interface{}{},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Even with allowConstructorFlags, invalid patterns should fail
			{
				Code: `RegExp('[', 'a');`,
				Options: map[string]interface{}{
					"allowConstructorFlags": []interface{}{"a"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Duplicate custom flags should still fail
			{
				Code: `RegExp('.', 'aa');`,
				Options: map[string]interface{}{
					"allowConstructorFlags": []interface{}{"a"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Flags not in allowConstructorFlags should fail
			{
				Code: `RegExp('.', 'z');`,
				Options: map[string]interface{}{
					"allowConstructorFlags": []interface{}{"a"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
		},
	)
}

func TestNoInvalidRegexpAdditionalCases(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidRegexpRule,
		[]rule_tester.ValidTestCase{
			// Character classes
			{Code: `new RegExp('[a-z]');`},
			{Code: `new RegExp('[0-9]');`},
			{Code: `new RegExp('[A-Za-z0-9]');`},

			// Quantifiers
			{Code: `new RegExp('a*');`},
			{Code: `new RegExp('a+');`},
			{Code: `new RegExp('a?');`},
			{Code: `new RegExp('a{1,3}');`},

			// Anchors
			{Code: `new RegExp('^abc$');`},
			{Code: `new RegExp('\\\\b\\\\w+\\\\b');`},

			// Groups
			{Code: `new RegExp('(a|b)');`},
			{Code: `new RegExp('(?:a|b)');`},

			// Lookaheads
			{Code: `new RegExp('a(?=b)');`},
			{Code: `new RegExp('a(?!b)');`},

			// Escape sequences
			{Code: `new RegExp('\\\\d+');`},
			{Code: `new RegExp('\\\\w+');`},
			{Code: `new RegExp('\\\\s+');`},
			{Code: `new RegExp('\\\\t\\\\n\\\\r');`},

			// Literal special characters
			{Code: `new RegExp('\\\\.');`},
			{Code: `new RegExp('\\\\*');`},
			{Code: `new RegExp('\\\\+');`},
			{Code: `new RegExp('\\\\?');`},
			{Code: `new RegExp('\\\\[');`},
			{Code: `new RegExp('\\\\]');`},
			{Code: `new RegExp('\\\\(');`},
			{Code: `new RegExp('\\\\)');`},
			{Code: `new RegExp('\\\\{');`},
			{Code: `new RegExp('\\\\}');`},
			{Code: `new RegExp('\\\\|');`},

			// Complex patterns
			{Code: `new RegExp('^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\\\.[a-zA-Z]{2,}$');`},
			{Code: `new RegExp('^\\\\d{3}-\\\\d{2}-\\\\d{4}$');`},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid ranges
			{
				Code: `new RegExp('[z-a]');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Invalid repetition patterns
			{
				Code: `new RegExp('**');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			{
				Code: `new RegExp('++');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Unclosed groups
			{
				Code: `new RegExp('(abc');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
			// Unclosed character class
			{
				Code: `new RegExp('[abc');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage"},
				},
			},
		},
	)
}
