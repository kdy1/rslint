package no_duplicate_case

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoDuplicateCaseRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDuplicateCaseRule,
		[]rule_tester.ValidTestCase{
			// Distinct numeric cases
			{Code: `switch (a) { case 1: break; case 2: break; }`},

			// Numeric vs string cases
			{Code: `switch (a) { case 1: break; case '1': break; }`},

			// Numeric vs boolean cases
			{Code: `switch (a) { case 1: break; case true: break; }`},

			// Only default clause
			{Code: `switch (a) { default: break; }`},

			// Different object property accesses
			{Code: `switch (a) { case p.p.p1: break; case p.p.p2: break; }`},

			// Function calls with different arguments
			{Code: `switch (a) { case f(true): break; case f(false): break; }`},

			// Different arithmetic expressions
			{Code: `switch (a) { case a + 1: break; case a + 2: break; }`},

			// Different ternary conditions
			{Code: `switch (a) { case a == 1 ? b : c: break; case a === 1 ? b : c: break; }`},

			// Calls to different functions
			{Code: `switch (a) { case f1(): break; case f2(): break; }`},

			// Array toString() conversions of different arrays
			{Code: `switch (a) { case [1].toString(): break; case [2].toString(): break; }`},

			// Separate switch statements with identical cases (allowed)
			{Code: `switch (a) { case 1: break; } switch (b) { case 1: break; }`},
		},
		[]rule_tester.InvalidTestCase{
			// Duplicate numeric literals
			{
				Code: `switch (a) { case 1: break; case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 34},
				},
			},

			// Duplicate string literals
			{
				Code: `switch (a) { case '1': break; case '1': break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36},
				},
			},

			// Duplicate variable references
			{
				Code: `var one = 1; switch (a) { case one: break; case one: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 49},
				},
			},

			// Duplicate object property chains
			{
				Code: `switch (a) { case p.p.p1: break; case p.p.p1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 39},
				},
			},

			// Duplicate function calls with same arguments
			{
				Code: `switch (a) { case f(true).p1: break; case f(true).p1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 43},
				},
			},

			// Duplicate expressions with same operations
			{
				Code: `switch (a) { case a + 1: break; case a + 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 38},
				},
			},

			// Duplicate ternary expressions
			{
				Code: `switch (a) { case a == 1 ? b : c: break; case a == 1 ? b : c: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 47},
				},
			},

			// Duplicate function calls from same function
			{
				Code: `switch (a) { case f(): break; case f(): break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36},
				},
			},

			// Duplicate array toString()
			{
				Code: `switch (a) { case [1].toString(): break; case [1].toString(): break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 47},
				},
			},

			// Fall-through duplicate cases without breaks
			{
				Code: `switch (a) { case 1: case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 27},
				},
			},

			// Multiple duplicates across single switch
			{
				Code: `switch (a) { case 1: break; case 2: break; case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 49},
				},
			},

			// Duplicates with whitespace variations (normalized)
			{
				Code: `switch (a) { case a+b: break; case a + b: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36},
				},
			},
		},
	)
}
