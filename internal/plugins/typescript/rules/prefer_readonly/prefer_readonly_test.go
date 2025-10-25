package prefer_readonly

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestPreferReadonlyRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferReadonlyRule,
		[]rule_tester.ValidTestCase{
			// Public property (not checked)
			{Code: `class Container { public prop = true; }`},
			// Protected property (not checked)
			{Code: `class Container { protected prop = true; }`},
			// Already readonly
			{Code: `class Container { private readonly prop = true; }`},
			// Modified in a method
			{Code: `class Container {
				private prop = true;
				mutate() {
					this.prop = false;
				}
			}`},
			// Parameter property already readonly
			{Code: `class Container {
				constructor(private readonly prop: string) {}
			}`},
		},
		[]rule_tester.InvalidTestCase{
			// Private property never modified
			{
				Code: `class Container { private neverModified = true; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferReadonly",
					},
				},
			},
			// Property only set in constructor
			{
				Code: `class Container {
					private onlyInConstructor: number;
					constructor(value: number) {
						this.onlyInConstructor = value;
					}
				}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferReadonly",
					},
				},
			},
			// Private parameter property never modified
			{
				Code: `class Container {
					constructor(private neverModified: string) {}
				}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferReadonly",
					},
				},
			},
		},
	)
}
