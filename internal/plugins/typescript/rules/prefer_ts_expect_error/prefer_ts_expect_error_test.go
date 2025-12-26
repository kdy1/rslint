package prefer_ts_expect_error

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferTsExpectErrorRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferTsExpectErrorRule,
		// Valid cases - should not trigger errors
		[]rule_tester.ValidTestCase{
			{Code: `// @ts-nocheck`},
			{Code: `// @ts-check`},
			{Code: `// just a comment containing @ts-ignore somewhere`},
			{Code: `
{
  /*
        just a comment containing @ts-ignore somewhere in a block
      */
}
			`},
			{Code: `// @ts-expect-error`},
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
		// Invalid cases - should trigger errors and be auto-fixable
		[]rule_tester.InvalidTestCase{
			{
				Code: `// @ts-ignore`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferExpectErrorComment", Line: 1, Column: 1},
				},
				Output: []string{`// @ts-expect-error`},
			},
			{
				Code: `// @ts-ignore: Suppress next line`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferExpectErrorComment", Line: 1, Column: 1},
				},
				Output: []string{`// @ts-expect-error: Suppress next line`},
			},
			{
				Code: `///@ts-ignore: Suppress next line`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferExpectErrorComment", Line: 1, Column: 1},
				},
				Output: []string{`///@ts-expect-error: Suppress next line`},
			},
			{
				Code: `
if (false) {
  // @ts-ignore: Unreachable code error
  console.log('hello');
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferExpectErrorComment", Line: 3, Column: 3},
				},
				Output: []string{`
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
				`},
			},
			{
				Code: `/* @ts-ignore */`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferExpectErrorComment", Line: 1, Column: 1},
				},
				Output: []string{`/* @ts-expect-error */`},
			},
			{
				Code: `
/**
 * Explaining comment
 *
 * @ts-ignore */
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferExpectErrorComment", Line: 2, Column: 1},
				},
				Output: []string{`
/**
 * Explaining comment
 *
 * @ts-expect-error */
				`},
			},
			{
				Code: `/* @ts-ignore in a single block */`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferExpectErrorComment", Line: 1, Column: 1},
				},
				Output: []string{`/* @ts-expect-error in a single block */`},
			},
			{
				Code: `
/*
// @ts-ignore in a block with single line comments */
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferExpectErrorComment", Line: 2, Column: 1},
				},
				Output: []string{`
/*
// @ts-expect-error in a block with single line comments */
				`},
			},
		},
	)
}
