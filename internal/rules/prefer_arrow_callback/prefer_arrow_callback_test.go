package prefer_arrow_callback

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestPreferArrowCallbackRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferArrowCallbackRule,
		[]rule_tester.ValidTestCase{
			// Arrow functions
			{Code: "foo(a => a);"},
			// Generator functions
			{Code: "foo(function*() {});"},
			// Functions using 'this' (with allowUnboundThis: true by default)
			{Code: "foo(function() { this.bar; });"},
			// Named functions with option
			{
				Code: "foo(function bar() {});",
				Options: map[string]interface{}{
					"allowNamedFunctions": true,
				},
			},
			// Self-referential functions
			{Code: "foo(function bar() { bar(); });"},
		},
		[]rule_tester.InvalidTestCase{
			// Basic conversion
			{
				Code: "foo(function() {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferArrow",
					},
				},
				Output: []string{"foo(() => {});"},
			},
			// Named function (without option)
			{
				Code: "foo(function bar() {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferArrow",
					},
				},
				Output: []string{"foo(() => {});"},
			},
			// With parameters
			{
				Code: "foo(function(a, b) { return a + b; });",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferArrow",
					},
				},
				Output: []string{"foo((a, b) => { return a + b; });"},
			},
			// Single parameter
			{
				Code: "foo(function(x) { return x * 2; });",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "preferArrow",
					},
				},
				Output: []string{"foo(x => { return x * 2; });"},
			},
		},
	)
}
