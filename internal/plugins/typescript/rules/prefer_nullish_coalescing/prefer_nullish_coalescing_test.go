package prefer_nullish_coalescing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferNullishCoalescingRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule, []rule_tester.ValidTestCase{
		// Valid cases - non-nullable types with ||
		{Code: `
declare let x: string;
(x || 'foo');
		`},
		{Code: `
declare let x: number;
(x || 5);
		`},
		{Code: `
declare let x: boolean;
(x || true);
		`},

		// Valid cases - nullable types with ??
		{Code: `
declare let x: string | null;
x ?? 'foo';
		`},
		{Code: `
declare let x: number | undefined;
x ?? 5;
		`},
		{Code: `
declare let x: boolean | null | undefined;
x ?? false;
		`},

		// Valid - ignoreTernaryTests option
		{
			Code: `
declare let x: string | undefined;
x !== undefined ? x : 'default';
			`,
			Options: PreferNullishCoalescingOptions{
				IgnoreTernaryTests: utils.Ref(true),
			},
		},

		// Valid - ignoreConditionalTests (default true)
		{Code: `
declare let x: string | null;
if (x || 'default') {}
		`},

		// Valid - any type
		{Code: `
declare let x: any;
x || 'foo';
		`},

	}, []rule_tester.InvalidTestCase{
		// Invalid - nullable type with ||
		{
			Code: `
declare let x: string | null;
x || 'foo';
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Line:      3,
				},
			},
		},
		{
			Code: `
declare let x: number | undefined;
x || 5;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Line:      3,
				},
			},
		},
		{
			Code: `
declare let x: string | null | undefined;
x || 'default';
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Line:      3,
				},
			},
		},

		// Invalid - ||= operator
		{
			Code: `
declare let x: string | undefined;
x ||= 'default';
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Line:      3,
				},
			},
		},

		// Invalid - ternary expression
		{
			Code: `
declare let x: string | null;
x !== null ? x : 'default';
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Line:      3,
				},
			},
		},
		{
			Code: `
declare let x: number | undefined;
x !== undefined ? x : 0;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Line:      3,
				},
			},
		},
	})
}
