package ban_ts_comment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBanTsComment(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		[]rule_tester.ValidTestCase{
			// No directive comments
			{Code: `const x = 1;`},
			{Code: `// Regular comment\nconst x = 1;`},

			// ts-expect-error with allow-with-description
			{
				Code: `// @ts-expect-error: TS2345 Argument type is incorrect\nconst x: number = "test";`,
				Options: map[string]interface{}{
					"ts-expect-error": "allow-with-description",
				},
			},

			// ts-check is allowed by default
			{Code: `// @ts-check\nconst x = 1;`},

			// Allow ts-ignore when configured
			{
				Code: `// @ts-ignore\nconst x = 1;`,
				Options: map[string]interface{}{
					"ts-ignore": false,
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Banned by default: ts-expect-error
			{
				Code: `// @ts-expect-error\nconst x: number = "test";`,
				Errors: []rule_tester.ExpectedError{
					{MessageId: "tsDirectiveComment"},
				},
			},

			// Banned by default: ts-ignore
			{
				Code: `// @ts-ignore\nconst x = 1;`,
				Errors: []rule_tester.ExpectedError{
					{MessageId: "tsDirectiveComment"},
				},
			},

			// Banned by default: ts-nocheck
			{
				Code: `// @ts-nocheck\nconst x = 1;`,
				Errors: []rule_tester.ExpectedError{
					{MessageId: "tsDirectiveComment"},
				},
			},

			// Description too short with allow-with-description
			{
				Code: `// @ts-expect-error: hi\nconst x: number = "test";`,
				Options: map[string]interface{}{
					"ts-expect-error":          "allow-with-description",
					"minimumDescriptionLength": 10,
				},
				Errors: []rule_tester.ExpectedError{
					{MessageId: "tsDirectiveCommentRequiresDescription"},
				},
			},

			// Block comment with ts-ignore
			{
				Code: `/* @ts-ignore */\nconst x = 1;`,
				Errors: []rule_tester.ExpectedError{
					{MessageId: "tsDirectiveComment"},
				},
			},
		},
		t,
		&BanTsCommentRule,
	)
}
