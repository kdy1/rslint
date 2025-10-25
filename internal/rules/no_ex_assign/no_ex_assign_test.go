package no_ex_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoExAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExAssignRule,
		[]rule_tester.ValidTestCase{
			// Basic catch with unrelated assignment
			{Code: `try { } catch (e) { three = 2 + 1; }`},

			// Destructured catch parameter
			{Code: `try { } catch ({e}) { this.something = 2; }`},

			// Function with catch block
			{Code: `function foo() { try { } catch (e) { return false; } }`},
		},
		[]rule_tester.InvalidTestCase{
			// Direct assignment to caught exception
			{
				Code: `try { } catch (e) { e = 10; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Assignment with different parameter name
			{
				Code: `try { } catch (ex) { ex = 10; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Array destructuring assignment
			{
				Code: `try { } catch (ex) { [ex] = []; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Object destructuring with default
			{
				Code: `try { } catch (ex) { ({x: ex = 0} = {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},

			// Destructured property assignment
			{
				Code: `try { } catch ({message}) { message = 10; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
		},
	)
}
