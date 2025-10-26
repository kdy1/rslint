package no_compare_neg_zero

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoCompareNegZeroRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoCompareNegZeroRule,
		[]rule_tester.ValidTestCase{
			{Code: `x === 0`},
			{Code: `0 === x`},
			{Code: `x == 0`},
			{Code: `0 == x`},
			{Code: `x === '0'`},
			{Code: `'0' === x`},
			{Code: `x == '0'`},
			{Code: `'0' == x`},
			{Code: `x === '-0'`},
			{Code: `'-0' === x`},
			{Code: `x == '-0'`},
			{Code: `'-0' == x`},
			{Code: `x === -1`},
			{Code: `-1 === x`},
			{Code: `x < 0`},
			{Code: `0 < x`},
			{Code: `x <= 0`},
			{Code: `0 <= x`},
			{Code: `x > 0`},
			{Code: `0 > x`},
			{Code: `x >= 0`},
			{Code: `0 >= x`},
			{Code: `x != 0`},
			{Code: `0 != x`},
			{Code: `x !== 0`},
			{Code: `0 !== x`},
			{Code: `Object.is(x, -0)`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `x === -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `-0 === x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `x == -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `-0 == x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `x > -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `-0 > x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `x >= -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `-0 >= x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `x < -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `-0 < x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `x <= -0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `-0 <= x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}
