package no_compare_neg_zero

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCompareNegZeroRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoCompareNegZeroRule,
		// Valid test cases
		[]rule_tester.ValidTestCase{
			{Code: "x === 0"},
			{Code: "0 === x"},
			{Code: "x == 0"},
			{Code: "0 == x"},
			{Code: "x === '0'"},
			{Code: "'0' === x"},
			{Code: "x == '0'"},
			{Code: "'0' == x"},
			{Code: "x === '-0'"},
			{Code: "'-0' === x"},
			{Code: "x == '-0'"},
			{Code: "'-0' == x"},
			{Code: "x === -1"},
			{Code: "-1 === x"},
			{Code: "x < 0"},
			{Code: "0 < x"},
			{Code: "x <= 0"},
			{Code: "0 <= x"},
			{Code: "x > 0"},
			{Code: "0 > x"},
			{Code: "x >= 0"},
			{Code: "0 >= x"},
			{Code: "x != 0"},
			{Code: "0 != x"},
			{Code: "x !== 0"},
			{Code: "0 !== x"},
			{Code: "Object.is(x, -0)"},
		},
		// Invalid test cases
		[]rule_tester.InvalidTestCase{
			{
				Code: "x === -0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "-0 === x",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "x == -0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "-0 == x",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "x > -0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "-0 > x",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "x >= -0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "-0 >= x",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "x < -0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "-0 < x",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "x <= -0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: "-0 <= x",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
		},
	)
}
