package no_class_assign

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoClassAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoClassAssignRule,
		[]rule_tester.ValidTestCase{
			{Code: `class A { } foo(A);`},
			{Code: `let A = class A { }; foo(A);`},
			{Code: `class A { b(A) { A = 0; } }`},
			{Code: `class A { b() { let A; A = 0; } }`},
			{Code: `let A = class { b() { A = 0; } }`},
			{Code: `var x = 0; x = 1;`},
			{Code: `let x = 0; x = 1;`},
			{Code: `const x = 0; x = 1;`},
			{Code: `function x() {} x = 1;`},
			{Code: `function foo(x) { x = 1; }`},
			{Code: `try {} catch (x) { x = 1; }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `class A { } A = 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "class",
						Line:      1,
						Column:    13,
					},
				},
			},
			{
				Code: `class A { } ({A} = 0);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "class",
						Line:      1,
						Column:    15,
					},
				},
			},
			{
				Code: `class A { } ({b: A = 0} = {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "class",
						Line:      1,
						Column:    18,
					},
				},
			},
			{
				Code: `A = 0; class A { }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "class",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `class A { b() { A = 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "class",
						Line:      1,
						Column:    17,
					},
				},
			},
			{
				Code: `let A = class A { b() { A = 0; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "class",
						Line:      1,
						Column:    25,
					},
				},
			},
			{
				Code: `class A { } A = 0; A = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "class",
						Line:      1,
						Column:    13,
					},
					{
						MessageId: "class",
						Line:      1,
						Column:    20,
					},
				},
			},
		},
	)
}
