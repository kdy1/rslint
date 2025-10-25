package no_class_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoClassAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoClassAssignRule,
		// Valid test cases
		[]rule_tester.ValidTestCase{
			{Code: "class A { }; foo(A);"},
			{Code: "let A = class A { }; foo(A);"},
			{Code: "class A { b(A) { A = 0; } }"},
			{Code: "class A { b() { let A; A = 0; } }"},
			{Code: "let A = class { b() { A = 0; } }"},
			{Code: "var x = 0;"},
			{Code: "let x = 0;"},
			{Code: "const x = 0;"},
			{Code: "function x() {}"},
			{Code: "var x; x = 0;"},
			{Code: "var x; ({ x } = 0);"},
			{Code: "var x; ({ y: x } = 0);"},
			{Code: "let x; x = 0;"},
			{Code: "let x; ({ x } = 0);"},
			{Code: "let x; ({ y: x } = 0);"},
		},
		// Invalid test cases
		[]rule_tester.InvalidTestCase{
			{
				Code: "class A { }; A = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "class"},
				},
			},
			{
				Code: "class A { }; ({A} = 0);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "class"},
				},
			},
			{
				Code: "class A { }; ({b: A = 0} = {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "class"},
				},
			},
			{
				Code: "A = 0; class A { }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "class"},
				},
			},
			{
				Code: "class A { b() { A = 0; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "class"},
				},
			},
			{
				Code: "let A = class A { b() { A = 0; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "class"},
				},
			},
			{
				Code: "class A { } A = 0; A = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "class"},
					{MessageId: "class"},
				},
			},
			{
				Code: "class A { } A += 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "class"},
				},
			},
			{
				Code: "class A { } A++;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "class"},
				},
			},
			{
				Code: "class A { } ++A;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "class"},
				},
			},
		},
	)
}
