package ban_ts_comment

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestBanTsCommentRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&BanTsCommentRule,
		[]rule_tester.ValidTestCase{
			// Regular comments without directives
			{Code: "// This is a normal comment\nconst x = 1;"},
			{Code: "/* This is a block comment */\nconst x = 1;"},

			// ts-check is allowed by default
			{Code: "// @ts-check\nconst x = 1;"},

			// ts-expect-error with description (allowed by default)
			{Code: "// @ts-expect-error: This is a description\nconst x: any = 1;"},
			{Code: "// @ts-expect-error - This is a description\nconst x: any = 1;"},

			// ts-nocheck with description (allowed by default)
			{Code: "// @ts-nocheck: This has a description\n"},

			// Mentioning directive in non-directive context
			{Code: "// This comment mentions ts-ignore but is not a directive\nconst x = 1;"},
		},
		[]rule_tester.InvalidTestCase{
			// ts-ignore is banned by default
			{
				Code: "// @ts-ignore\nconst x = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tsDirectiveComment", Line: 1, Column: 1},
				},
			},

			// ts-expect-error without description
			{
				Code: "// @ts-expect-error\nconst x = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tsDirectiveCommentRequiresDescription", Line: 1, Column: 1},
				},
			},

			// ts-nocheck without description
			{
				Code: "// @ts-nocheck\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tsDirectiveCommentRequiresDescription", Line: 1, Column: 1},
				},
			},

			// ts-expect-error with too short description
			{
				Code: "// @ts-expect-error: a\nconst x = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern", Line: 1, Column: 1},
				},
			},

			// Block comment with ts-ignore
			{
				Code: "/* @ts-ignore */\nconst x = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "tsDirectiveComment", Line: 1, Column: 1},
				},
			},
		},
	)
}
