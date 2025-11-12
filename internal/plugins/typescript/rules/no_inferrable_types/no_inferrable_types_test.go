package no_inferrable_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInferrableTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(t, NoInferrableTypesRule, []rule_tester.RuleTestCase{
		// Valid cases - no type annotations
		{Code: "const a = 10n;"},
		{Code: "const a = false;"},
		{Code: "const a = true;"},
		{Code: "const a = 10;"},
		{Code: "const a = 'str';"},
		{Code: "const a = `str`;"},
		{Code: "const a = null;"},
		{Code: "const a = undefined;"},
		{Code: "const a = /a/;"},

		// Valid cases - with "any" type (explicitly allowed)
		{Code: "const a: any = 5;"},

		// Valid cases - ignoreParameters option
		{
			Code:    "const fn = (a: number = 5) => {};",
			Options: []interface{}{map[string]interface{}{"ignoreParameters": true}},
		},
		{
			Code:    "function fn(a: number = 5) {}",
			Options: []interface{}{map[string]interface{}{"ignoreParameters": true}},
		},

		// Valid cases - ignoreProperties option
		{
			Code: `
class Foo {
  a: number = 5;
}
			`,
			Options: []interface{}{map[string]interface{}{"ignoreProperties": true}},
		},

		// Valid cases - optional properties (always allowed)
		{
			Code: `
class Foo {
  a?: number = 5;
}
			`,
		},

		// Valid cases - readonly properties (always allowed)
		{
			Code: `
class Foo {
  readonly a: number = 5;
}
			`,
		},

		// Invalid cases - bigint
		{
			Code: "const a: bigint = 10n;",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = 10n;",
		},
		{
			Code: "const a: bigint = -10n;",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = -10n;",
		},

		// Invalid cases - boolean
		{
			Code: "const a: boolean = false;",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = false;",
		},
		{
			Code: "const a: boolean = true;",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = true;",
		},
		{
			Code: "const a: boolean = !0;",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = !0;",
		},

		// Invalid cases - number
		{
			Code: "const a: number = 10;",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = 10;",
		},
		{
			Code: "const a: number = +10;",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = +10;",
		},
		{
			Code: "const a: number = -10;",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = -10;",
		},

		// Invalid cases - string
		{
			Code: "const a: string = 'str';",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = 'str';",
		},
		{
			Code: "const a: string = `str`;",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = `str`;",
		},

		// Invalid cases - null
		{
			Code: "const a: null = null;",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = null;",
		},

		// Invalid cases - undefined
		{
			Code: "const a: undefined = undefined;",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const a = undefined;",
		},

		// Invalid cases - function parameters
		{
			Code: "const fn = (a: number = 5) => {};",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "const fn = (a = 5) => {};",
		},
		{
			Code: "function fn(a: number = 5) {}",
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: "function fn(a = 5) {}",
		},

		// Invalid cases - class properties
		{
			Code: `
class Foo {
  a: number = 5;
}
			`,
			Errors: []rule_tester.ErrorCase{
				{MessageId: "noInferrableType"},
			},
			Output: `
class Foo {
  a = 5;
}
			`,
		},
	})
}
