package no_confusing_non_null_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConfusingNonNullAssertionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoConfusingNonNullAssertionRule,
		[]rule_tester.ValidTestCase{
			// Valid cases - non-null assertion not in confusing positions
			{Code: `a == b!;`},
			{Code: `a = b!;`},
			{Code: `a !== b;`},
			{Code: `a != b;`},
			{Code: `(a + b!) == c;`},
			{Code: `(a + b!) = c;`},
			{Code: `(a + b!) in c;`},
			{Code: `(a || b!) instanceof c;`},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid cases - equality operators
			{
				Code: `a! == b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "confusingEqual",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 3,
					},
				},
			},
			{
				Code: `a! === b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "confusingEqual",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 3,
					},
				},
			},
			{
				Code: `a + b! == c;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "confusingEqual",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 13,
					},
				},
			},
			{
				Code: `(obj = new new OuterObj().InnerObj).Name! == c;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "confusingEqual",
						Line:      1,
						Column:    37,
						EndLine:   1,
						EndColumn: 46,
					},
				},
			},
			{
				Code: `(a==b)! ==c;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "confusingEqual",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 8,
					},
				},
			},
			// Invalid cases - assignment operator
			{
				Code: `a! = b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "confusingAssign",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 3,
					},
				},
			},
			{
				Code: `(obj = new new OuterObj().InnerObj).Name! = c;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "confusingAssign",
						Line:      1,
						Column:    37,
						EndLine:   1,
						EndColumn: 46,
					},
				},
			},
			{
				Code: `(a=b)! =c;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "confusingAssign",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 7,
					},
				},
			},
			// Invalid cases - 'in' operator
			{
				Code: `a! in b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "confusingOperator",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 3,
					},
				},
			},
			{
				Code: `
a!
  in b;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "confusingOperator",
						Line:      2,
						Column:    1,
						EndLine:   2,
						EndColumn: 3,
					},
				},
			},
			// Invalid cases - 'instanceof' operator
			{
				Code: `a! instanceof b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "confusingOperator",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 3,
					},
				},
			},
		},
	)
}
