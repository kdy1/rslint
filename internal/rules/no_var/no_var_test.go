package no_var

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoVarRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoVarRule,
		[]rule_tester.ValidTestCase{
			// const declarations
			{Code: "const JOE = 'schmoe';"},
			// let declarations
			{Code: "let moo = 'car';"},
			// using declarations (ES2026)
			{Code: "using moo = 'car';"},
			// await using declarations
			{Code: "await using moo = 'car';"},
		},
		[]rule_tester.InvalidTestCase{
			// Simple var replacement
			{
				Code: "var foo = bar;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedVar",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"let foo = bar;"},
			},
			// Multiple declarations
			{
				Code: "var foo = bar, toast = most;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedVar",
						Line:      1,
						Column:    1,
					},
				},
				Output: []string{"let foo = bar, toast = most;"},
			},
			// For loop
			{
				Code: "for (var i = 0; i < list.length; ++i) { foo(i) }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedVar",
					},
				},
				Output: []string{"for (let i = 0; i < list.length; ++i) { foo(i) }"},
			},
			// Destructuring
			{
				Code: "var {a, b} = obj;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedVar",
					},
				},
				Output: []string{"let {a, b} = obj;"},
			},
			// Array destructuring
			{
				Code: "var [a, b] = arr;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedVar",
					},
				},
				Output: []string{"let [a, b] = arr;"},
			},
		},
	)
}
