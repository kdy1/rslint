package no_empty

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyRule,
		[]rule_tester.ValidTestCase{
			// Non-empty blocks are valid
			{Code: `if (foo) { bar() }`},
			{Code: `while (foo) { bar() }`},
			{Code: `for (;;) { bar() }`},
			{Code: `try { foo() } catch (ex) { bar() }`},
			{Code: `switch(foo) { case bar: break; }`},
			{Code: `(function() { foo(); }())`},
			// Empty blocks with comments are valid
			{Code: `if (foo) {/* empty */}`},
			{Code: `while (foo) {/* empty */}`},
			{Code: `try {/* empty */} catch (ex) {foo()}`},
			{Code: `if (foo) { // empty }`},
		},
		[]rule_tester.InvalidTestCase{
			// Empty blocks without comments are invalid
			{
				Code: `try {} catch (ex) {throw ex}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `if (foo) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `while (foo) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `for (;foo;) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `switch(foo) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
		},
	)
}

func TestNoEmptyRuleWithOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyRule,
		[]rule_tester.ValidTestCase{
			// Empty catch blocks are allowed with allowEmptyCatch option
			{
				Code: `try { foo() } catch (ex) {}`,
				Options: map[string]interface{}{
					"allowEmptyCatch": true,
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Empty catch blocks still invalid without option
			{
				Code: `try { foo() } catch (ex) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
		},
	)
}
