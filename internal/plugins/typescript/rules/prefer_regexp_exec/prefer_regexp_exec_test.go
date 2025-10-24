package prefer_regexp_exec

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferRegexpExecRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferRegexpExecRule,
		[]rule_tester.ValidTestCase{
			{Code: `
const result = /foo/.exec("bar");
`},
			{Code: `
const result = "foo".match(/bar/g);
`},
			{Code: `
const regex = /test/g;
const result = "string".match(regex);
`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
const result = "foo".match(/bar/);
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "regExpExecOverStringMatch",
						Line:      2,
						Column:    16,
					},
				},
			},
			{
				Code: `
const regex = /test/;
const result = "string".match(regex);
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "regExpExecOverStringMatch",
						Line:      3,
						Column:    16,
					},
				},
			},
		},
	)
}
