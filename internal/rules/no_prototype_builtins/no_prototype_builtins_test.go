package no_prototype_builtins

import (
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
	"testing"
)

func TestNoPrototypeBuiltinsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoPrototypeBuiltinsRule,
		[]rule_tester.ValidTestCase{
			// Using Object.prototype with call() - recommended approach
			{Code: "Object.prototype.hasOwnProperty.call(foo, 'bar')"},
			{Code: "Object.prototype.isPrototypeOf.call(foo, 'bar')"},
			{Code: "Object.prototype.propertyIsEnumerable.call(foo, 'bar')"},
			{Code: "Object.prototype.hasOwnProperty.apply(foo, ['bar'])"},

			// Using empty object with call/apply
			{Code: "({}.hasOwnProperty.call(foo, 'bar'))"},
			{Code: "({}.isPrototypeOf.apply(foo, ['bar']))"},

			// Not method calls
			{Code: "foo.hasOwnProperty"},
			{Code: "foo.hasOwnProperty.bar()"},
			{Code: "foo(hasOwnProperty)"},
			{Code: "hasOwnProperty(foo, 'bar')"},

			// Bracket notation with non-matching names
			{Code: "foo[hasOwnProperty]('bar')"},
			{Code: "foo['HasOwnProperty']('bar')"},
			{Code: "foo[1]('bar')"},
			{Code: "foo[null]('bar')"},
		},
		[]rule_tester.InvalidTestCase{
			// Basic violations with dot notation
			{
				Code: "foo.hasOwnProperty('bar')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn", Line: 1, Column: 1},
				},
				Output: []string{"Object.prototype.hasOwnProperty.call(foo, 'bar')"},
			},
			{
				Code: "foo.isPrototypeOf('bar')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn", Line: 1, Column: 1},
				},
				Output: []string{"Object.prototype.isPrototypeOf.call(foo, 'bar')"},
			},
			{
				Code: "foo.propertyIsEnumerable('bar')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn", Line: 1, Column: 1},
				},
				Output: []string{"Object.prototype.propertyIsEnumerable.call(foo, 'bar')"},
			},

			// Nested property access
			{
				Code: "foo.bar.hasOwnProperty('bar')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn", Line: 1, Column: 1},
				},
				Output: []string{"Object.prototype.hasOwnProperty.call(foo.bar, 'bar')"},
			},
			{
				Code: "foo.bar.baz.isPrototypeOf('bar')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn", Line: 1, Column: 1},
				},
				Output: []string{"Object.prototype.isPrototypeOf.call(foo.bar.baz, 'bar')"},
			},

			// Bracket notation with string literal
			{
				Code: "foo['hasOwnProperty']('bar')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn", Line: 1, Column: 1},
				},
				Output: []string{"Object.prototype.hasOwnProperty.call(foo, 'bar')"},
			},
			{
				Code: "foo.bar[\"propertyIsEnumerable\"]('baz')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn", Line: 1, Column: 1},
				},
				Output: []string{"Object.prototype.propertyIsEnumerable.call(foo.bar, 'baz')"},
			},

			// Multiple arguments
			{
				Code: "foo.hasOwnProperty('bar', 'baz')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prototypeBuildIn", Line: 1, Column: 1},
				},
				Output: []string{"Object.prototype.hasOwnProperty.call(foo, 'bar', 'baz')"},
			},
		},
	)
}
