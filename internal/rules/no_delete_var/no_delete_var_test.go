package no_delete_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDeleteVarRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDeleteVarRule,
		[]rule_tester.ValidTestCase{
			// Deleting object properties is allowed
			{Code: `delete x.prop;`},
		},
		[]rule_tester.InvalidTestCase{
			// Deleting variable references is not allowed
			{
				Code: `delete x`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
					},
				},
			},
		},
	)
}
