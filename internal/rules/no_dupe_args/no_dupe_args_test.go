package no_dupe_args

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeArgsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeArgsRule,
		[]rule_tester.ValidTestCase{
			// Valid cases - no duplicate parameters
			{Code: `function a(a, b, c){}`},
			{Code: `var a = function(a, b, c){}`},
			{Code: `function a({a, b}, {c, d}){}`},
			{Code: `function a([ , a]) {}`},
			{Code: `function foo([[a, b], [c, d]]) {}`},
			{Code: `function a({a, b}){}`},
			{Code: `var a = function({a, b, c}){}`},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid cases - duplicate parameters
			{
				Code: `function a(a, b, b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "b",
						},
					},
				},
			},
			{
				Code: `function a(a, a, a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "a",
						},
					},
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "a",
						},
					},
				},
			},
			{
				Code: `function a(a, b, a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "a",
						},
					},
				},
			},
			{
				Code: `function a(a, b, a, b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "a",
						},
					},
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "b",
						},
					},
				},
			},
			{
				Code: `var a = function(a, b, b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "b",
						},
					},
				},
			},
			{
				Code: `var a = function(a, a, a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "a",
						},
					},
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "a",
						},
					},
				},
			},
			{
				Code: `var a = function(a, b, a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "a",
						},
					},
				},
			},
			{
				Code: `var a = function(a, b, a, b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "a",
						},
					},
					{
						MessageId: "unexpected",
						Data: map[string]interface{}{
							"name": "b",
						},
					},
				},
			},
		},
	)
}
