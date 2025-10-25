package no_dupe_class_members

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDupeClassMembersRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDupeClassMembersRule,
		[]rule_tester.ValidTestCase{
			// Valid cases - no duplicate class members
			{Code: "class A { foo() {} bar() {} }"},
			{Code: "class A { foo() {} static foo() {} }"},
			{Code: "class A { get foo() {} set foo(value) {} }"},
			{Code: "class A { foo() {} } class B { foo() {} }"},
			{Code: "class A { [foo]() {} [bar]() {} }"},
			{Code: "class A { 10() {} 20() {} }"},
			{Code: "class A { ['foo']() {} ['bar']() {} }"},
			{Code: "class A { constructor() {} foo() {} }"},
			{Code: "class A { foo; bar; }"},
			{Code: "class A { foo; static foo; }"},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid cases - duplicate class members
			{
				Code: "class A { foo() {} foo() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateMember",
						Line:      1,
						Column:    20,
					},
				},
			},
			{
				Code: "class A { static foo() {} static foo() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateMember",
						Line:      1,
						Column:    27,
					},
				},
			},
			{
				Code: "class A { foo() {} foo() {} foo() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateMember",
						Line:      1,
						Column:    20,
					},
					{
						MessageId: "duplicateMember",
						Line:      1,
						Column:    29,
					},
				},
			},
			{
				Code: "class A { get foo() {} foo() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateMember",
						Line:      1,
						Column:    24,
					},
				},
			},
			{
				Code: "class A { foo() {} set foo(value) {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateMember",
						Line:      1,
						Column:    20,
					},
				},
			},
			{
				Code: "class A { foo; foo; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateMember",
						Line:      1,
						Column:    16,
					},
				},
			},
			{
				Code: "class A { foo; foo() {} }",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "duplicateMember",
						Line:      1,
						Column:    16,
					},
				},
			},
		},
	)
}
