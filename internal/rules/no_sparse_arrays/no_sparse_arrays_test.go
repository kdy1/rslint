package no_sparse_arrays

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSparseArraysRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoSparseArraysRule,
		[]rule_tester.ValidTestCase{
			{Code: `var a = [ 1, 2, ]`},
			{Code: `var a = [1, 2, 3]`},
			{Code: `var a = []`},
			{Code: `var a = [1,]`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `var a = [,];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedSparseArray",
						Line:      1,
						Column:    10,
					},
				},
			},
			{
				Code: `var a = [ 1,, 2];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedSparseArray",
						Line:      1,
						Column:    13,
					},
				},
			},
			{
				Code: `var a = [1, , 3]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedSparseArray",
						Line:      1,
						Column:    13,
					},
				},
			},
		},
	)
}
