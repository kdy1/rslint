package no_loss_of_precision

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoLossOfPrecisionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoLossOfPrecisionRule,
		[]rule_tester.ValidTestCase{
			// Simple numbers
			{Code: `var x = 12345;`},
			{Code: `var x = 123.456;`},
			{Code: `var x = -123.456;`},
			{Code: `var x = 0;`},

			// Scientific notation
			{Code: `var x = 123e34;`},
			{Code: `var x = 123e-34;`},
			{Code: `var x = 12.3e-34;`},

			// MAX_SAFE_INTEGER
			{Code: `var x = 9007199254740991;`},

			// Large safe integers
			{Code: `var x = 12300000000000000000000000;`},

			// Small decimals
			{Code: `var x = 0.00000000000000000000000123;`},

			// Binary, octal, hex
			{Code: `var x = 0b11111111111111111111111111111111111111111111111111111;`},
			{Code: `var x = 0o377777777777777777;`},
			{Code: `var x = 0x1FFFFFFFFFFFFF;`},

			// Non-numeric
			{Code: `var x = true;`},
			{Code: `var x = 'abc';`},
			{Code: `var x = null;`},
		},
		[]rule_tester.InvalidTestCase{
			// Integers with precision loss
			{
				Code: `var x = 9007199254740993;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLossOfPrecision"},
				},
			},
			{
				Code: `var x = 5123000000000000000000000000001;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLossOfPrecision"},
				},
			},

			// Scientific notation with precision loss
			{
				Code: `var x = 9007199254740.993e3;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLossOfPrecision"},
				},
			},
			{
				Code: `var x = 9.007199254740993e15;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLossOfPrecision"},
				},
			},

			// Decimals with precision loss
			{
				Code: `var x = 900719.9254740994;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLossOfPrecision"},
				},
			},

			// Hex with precision loss
			{
				Code: `var x = 0x20000000000001;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLossOfPrecision"},
				},
			},

			// Binary with precision loss
			{
				Code: `var x = 0b100000000000000000000000000000000000000000000000000001;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noLossOfPrecision"},
				},
			},
		},
	)
}
