package prefer_function_type

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestPreferFunctionTypeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferFunctionTypeRule,
		[]rule_tester.ValidTestCase{
			// Interface with call signature and properties
			{Code: `interface Foo { (): void; bar: number; }`},
			// Type literal with call signature and properties
			{Code: `type Foo = { (): void; bar: number; }`},
			// Extended interface
			{Code: `interface Bar extends Foo { (): void; }`},
			// Already a function type
			{Code: `type Foo = () => string;`},
		},
		[]rule_tester.InvalidTestCase{
			// Simple interface with only call signature
			{
				Code: `interface Foo { (): string; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "functionTypeOverCallableType",
					},
				},
			},
			// Type literal with only call signature
			{
				Code: `type Foo = { (): string; };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "functionTypeOverCallableType",
					},
				},
			},
			// Interface with generic and call signature
			{
				Code: `interface Foo<T> { (bar: T): string; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "functionTypeOverCallableType",
					},
				},
			},
			// Interface with this parameter (should report different error)
			{
				Code: `interface Foo { (arg: this): void; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedThisOnFunctionOnlyInterface",
					},
				},
			},
		},
	)
}
