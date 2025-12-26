package prefer_ts_expect_error_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/prefer_ts_expect_error"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferTsExpectError(t *testing.T) {
	tester := rule_tester.NewRuleTester(t)

	tester.Run("prefer-ts-expect-error", prefer_ts_expect_error.PreferTsExpectErrorRule, &rule_tester.RuleTesterConfig{
		Valid: []rule_tester.ValidTestCase{
			{Code: "// @ts-nocheck"},
			{Code: "// @ts-check"},
			{Code: "// just a comment containing @ts-ignore somewhere"},
			{Code: `
{
  /*
    just a comment containing @ts-ignore somewhere in a block
  */
}
			`},
			{Code: "// @ts-expect-error"},
			{Code: `
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
			`},
			{Code: `
/**
 * Explaining comment
 *
 * @ts-expect-error
 *
 * Not last line
 * */
			`},
		},
		Invalid: []rule_tester.InvalidTestCase{
			{
				Code: "// @ts-ignore",
				Errors: []rule_tester.ExpectedError{
					{
						Column:    1,
						Line:      1,
						MessageId: "preferExpectErrorComment",
					},
				},
				Output: "// @ts-expect-error",
			},
			{
				Code: "// @ts-ignore: Suppress next line",
				Errors: []rule_tester.ExpectedError{
					{
						Column:    1,
						Line:      1,
						MessageId: "preferExpectErrorComment",
					},
				},
				Output: "// @ts-expect-error: Suppress next line",
			},
			{
				Code: "///@ts-ignore: Suppress next line",
				Errors: []rule_tester.ExpectedError{
					{
						Column:    1,
						Line:      1,
						MessageId: "preferExpectErrorComment",
					},
				},
				Output: "///@ts-expect-error: Suppress next line",
			},
			{
				Code: `
if (false) {
  // @ts-ignore: Unreachable code error
  console.log('hello');
}
				`,
				Errors: []rule_tester.ExpectedError{
					{
						Column:    3,
						Line:      3,
						MessageId: "preferExpectErrorComment",
					},
				},
				Output: `
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
				`,
			},
			{
				Code: "/* @ts-ignore */",
				Errors: []rule_tester.ExpectedError{
					{
						Column:    1,
						Line:      1,
						MessageId: "preferExpectErrorComment",
					},
				},
				Output: "/* @ts-expect-error */",
			},
			{
				Code: `
/**
 * Explaining comment
 *
 * @ts-ignore */
				`,
				Errors: []rule_tester.ExpectedError{
					{
						Column:    1,
						Line:      2,
						MessageId: "preferExpectErrorComment",
					},
				},
				Output: `
/**
 * Explaining comment
 *
 * @ts-expect-error */
				`,
			},
			{
				Code: "/* @ts-ignore in a single block */",
				Errors: []rule_tester.ExpectedError{
					{
						Column:    1,
						Line:      1,
						MessageId: "preferExpectErrorComment",
					},
				},
				Output: "/* @ts-expect-error in a single block */",
			},
			{
				Code: `
/*
// @ts-ignore in a block with single line comments */
				`,
				Errors: []rule_tester.ExpectedError{
					{
						Column:    1,
						Line:      2,
						MessageId: "preferExpectErrorComment",
					},
				},
				Output: `
/*
// @ts-expect-error in a block with single line comments */
				`,
			},
		},
	})
}
