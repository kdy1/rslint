package no_unnecessary_parameter_property_assignment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryParameterPropertyAssignmentRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryParameterPropertyAssignmentRule,
		[]rule_tester.ValidTestCase{
			{Code: `
class C {
  constructor(public x: number) {
    // No redundant assignment
  }
}
`},
			{Code: `
class C {
  x: number;
  constructor(x: number) {
    this.x = x; // Valid - not a parameter property
  }
}
`},
		},
		// TODO: Add invalid test cases once rule implementation is complete
		[]rule_tester.InvalidTestCase{},
	)
}
