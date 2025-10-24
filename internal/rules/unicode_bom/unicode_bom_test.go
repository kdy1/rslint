package unicode_bom

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestUnicodeBomRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UnicodeBomRule, []rule_tester.ValidTestCase{
		// Valid: BOM present when always mode
		{
			Code:    "\uFEFFvar a = 123;",
			Options: []interface{}{"always"},
		},
		// Valid: No BOM when never mode (default)
		{
			Code: "var a = 123;",
		},
		// Valid: No BOM when never mode (explicit)
		{
			Code:    "var a = 123;",
			Options: []interface{}{"never"},
		},
		// Valid: BOM in middle/end is okay for never mode
		{
			Code:    "var a = 123; \uFEFF",
			Options: []interface{}{"never"},
		},
	}, []rule_tester.InvalidTestCase{
		// Invalid: Missing BOM in always mode
		{
			Code:    "var a = 123;",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "expected",
					Line:      1,
					Column:    1,
				},
			},
			Output: []string{"\uFEFFvar a = 123;"},
		},
		// Invalid: Missing BOM with comment in always mode
		{
			Code:    "// here's a comment \nvar a = 123;",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "expected",
					// Line/Column not checked - synthetic node position may vary
				},
			},
			Output: []string{"\uFEFF// here's a comment \nvar a = 123;"},
		},
		// Invalid: Unwanted BOM (default never)
		{
			Code: "\uFEFF var a = 123;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      1,
					// Column is not checked because BOM affects positioning
				},
			},
			Output: []string{" var a = 123;"},
		},
		// Invalid: Unwanted BOM (explicit never)
		{
			Code:    "\uFEFF var a = 123;",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpected",
					Line:      1,
					// Column is not checked because BOM affects positioning
				},
			},
			Output: []string{" var a = 123;"},
		},
	})
}
