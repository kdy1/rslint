package no_new_native_nonconstructor

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoNewNativeNonconstructorRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewNativeNonconstructorRule,
		[]rule_tester.ValidTestCase{
			// Direct function calls (correct usage)
			{Code: `var foo = Symbol('foo');`},
			{Code: `var foo = BigInt(9007199254740991);`},

			// Symbol/BigInt as parameters (shadowing)
			{Code: `function bar(Symbol) { var baz = new Symbol('baz'); }`},
			{Code: `function bar(BigInt) { var baz = new BigInt(9007199254740991); }`},

			// Local function declarations
			{Code: `function Symbol() {} new Symbol();`},
			{Code: `function BigInt() {} new BigInt();`},

			// As arguments
			{Code: `new foo(Symbol);`},
			{Code: `new foo(bar, BigInt);`},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid: new Symbol
			{
				Code: `var foo = new Symbol('foo');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noNewNativeNonconstructor",
					},
				},
			},

			// Invalid: new BigInt
			{
				Code: `var foo = new BigInt(9007199254740991);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noNewNativeNonconstructor",
					},
				},
			},
		},
	)
}
