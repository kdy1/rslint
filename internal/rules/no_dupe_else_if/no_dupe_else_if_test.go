package no_dupe_else_if

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeElseIfRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeElseIfRule,
		[]rule_tester.ValidTestCase{
			// Valid cases - different conditions
			{Code: `if (a) {} else if (b) {}`},
			{Code: `if (a === 1) {} else if (a === 2) {}`},
			{Code: `if (f(a)) {} else if (g(a)) {}`},
			{Code: `if (a) {}`},
			{Code: `if (a) {} else {}`},
			{Code: `if (a) {} else if (b) {} else {}`},
			{Code: `if (a) {} else if (b) {} else if (c) {}`},
			{Code: `if (a || b) {} else if (c || d) {}`},
			{Code: `if (a || b) {} else if (a || c) {}`},
			{Code: `if (a && b) {} else if (a) {} else if (b) {}`},
			{Code: `if (a) { if (a) {} }`},
			{Code: `if (a) {} if (a) {}`},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid cases - duplicate conditions
			{
				Code: `if (a) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (a) {} else if (a) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (a === 1) {} else if (a === 1) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (a && b) {} else if (a && b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (f(a)) {} else if (f(a)) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (a || b) {} else if (a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (a) {} else if (a || b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (a) {} else if (b) {} else if (a || b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (a || b) {} else if (b && c) {} else if (a || b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
		},
	)
}
