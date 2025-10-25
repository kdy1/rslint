package no_self_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSelfAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoSelfAssignRule,
		[]rule_tester.ValidTestCase{
			// Simple valid assignments
			{Code: "var a = a;"},         // Variable declaration (not assignment)
			{Code: "a = b;"},              // Different variables
			{Code: "a += a;"},             // Compound assignment
			{Code: "a = +a;"},             // Unary operator
			{Code: "a = [a];"},            // Array wrapping
			{Code: "a &= a;"},             // Bitwise compound
			{Code: "a |= a;"},             // Bitwise compound
			{Code: "[a] = a;"},            // Destructuring from non-array
			{Code: "[a = 1] = [a];"},      // With default value
			{Code: "[a, b] = [b, a];"},    // Swap pattern
			{Code: "[a,, b] = [, b, a];"}, // With holes
			{Code: "({a} = a);"},          // Object destructuring from non-object
			{Code: "({a = 1} = {a});"},    // With default value
			{Code: "({a: b} = {a});"},     // Different property name

			// Property access with props: false
			{Code: "a.b = a.b;", Options: map[string]interface{}{"props": false}},
			{Code: "a[b] = a[b];", Options: map[string]interface{}{"props": false}},

			// Different property access
			{Code: "a.b = a.c;"},
			{Code: "a.b = c.b;"},
		},
		[]rule_tester.InvalidTestCase{
			// Simple self-assignment
			{
				Code: "a = a;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment"},
				},
			},

			// Array destructuring
			{
				Code: "[a] = [a];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment"},
				},
			},
			{
				Code: "[a, b] = [a, b];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment"},
					{MessageId: "selfAssignment"},
				},
			},

			// Object destructuring
			{
				Code: "({a} = {a});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment"},
				},
			},
			{
				Code: "({a, b} = {a, b});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment"},
					{MessageId: "selfAssignment"},
				},
			},

			// Property access (props: true by default)
			{
				Code: "a.b = a.b;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment"},
				},
			},
			{
				Code: "a['b'] = a['b'];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment"},
				},
			},
			{
				Code: "a[b] = a[b];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment"},
				},
			},

			// Logical assignment operators (ES2021)
			{
				Code: "a &&= a;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment"},
				},
			},
			{
				Code: "a ||= a;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment"},
				},
			},
			{
				Code: "a ??= a;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment"},
				},
			},
		})
}
