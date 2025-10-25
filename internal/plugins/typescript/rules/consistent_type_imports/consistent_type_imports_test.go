package consistent_type_imports

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestConsistentTypeImportsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentTypeImportsRule,
		[]rule_tester.ValidTestCase{
			// TODO: Add valid test cases when rule is fully implemented
			{Code: `const x = 1;`},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases when rule is fully implemented
		},
	)
}

func TestConsistentTypeImportsRuleWithOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentTypeImportsRule,
		[]rule_tester.ValidTestCase{
			// TODO: Add valid test cases with options when rule is fully implemented
			{Code: `const x = 1;`},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases with options when rule is fully implemented
		},
	)
}
